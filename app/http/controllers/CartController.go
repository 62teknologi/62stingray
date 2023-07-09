package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/62teknologi/62stingray/62golib/utils"
	"github.com/62teknologi/62stingray/config"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type CartController struct {
	SingularName  string
	PluralName    string
	SingularLabel string
	PluralLabel   string
	Table         string
}

func (ctrl *CartController) Init(ctx *gin.Context) {
	ctrl.SingularName = utils.Pluralize.Singular(ctx.Param("table"))
	ctrl.PluralName = utils.Pluralize.Plural(ctx.Param("table"))
	ctrl.SingularLabel = ctrl.SingularName
	ctrl.PluralLabel = ctrl.PluralName
	ctrl.Table = ctrl.PluralName
}

func (ctrl CartController) FindAll(ctx *gin.Context) {
	ctrl.Init(ctx)

	values := []map[string]any{}
	columns := []string{ctrl.Table + ".*"}
	transformer, _ := utils.JsonFileParser(config.Data.SettingPath + "/transformers/response/" + ctrl.Table + "/find.json")
	query := utils.DB.Table(ctrl.Table)
	filter := utils.SetFilterByQuery(query, transformer, ctx)
	search := utils.SetGlobalSearch(query, transformer, ctx)

	utils.SetOrderByQuery(query, ctx)
	utils.SetBelongsTo(query, transformer, &columns)
	utils.SetOperation(query, transformer, &columns)

	delete(transformer, "filterable")
	delete(transformer, "searchable")

	pagination := utils.SetPagination(query, ctx)

	if err := query.Select(columns).Find(&values).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", ctrl.PluralLabel+" not found", nil))
		return
	}

	customResponses := utils.MultiMapValuesShifter(transformer, values)
	summary := utils.GetSummary(transformer, values)

	ctx.JSON(http.StatusOK, utils.ResponseDataPaginate("success", "find "+ctrl.PluralLabel+" success", customResponses, pagination, filter, search, summary))
}

func (ctrl CartController) Add(ctx *gin.Context) {
	ctrl.Init(ctx)

	transformer, _ := utils.JsonFileParser(config.Data.SettingPath + "/transformers/request/" + ctrl.Table + "/add.json")
	input := utils.ParseForm(ctx)

	if validation, err := utils.Validate(input, transformer); err {
		ctx.JSON(http.StatusOK, utils.ResponseData("failed", "validation", validation.Errors))
		return
	}

	utils.MapValuesShifter(transformer, input)
	utils.MapNullValuesRemover(transformer)

	setting := transformer["setting"].(map[string]any)
	userField := setting["user_id"].(string)
	productField := setting["product_id"].(string)
	quantityField := setting["quantity"].(string)

	delete(transformer, "setting")

	query := utils.DB.Table(ctrl.Table).Where(userField+"= ?", transformer[userField]).Where(productField+"= ?", transformer[productField]).Updates(map[string]any{
		quantityField: 1, //needed to trigger changes
		quantityField: gorm.Expr(quantityField+" + ?", 1),
	})

	if query.Error != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", query.Error.Error(), nil))
		return
	}

	if rowsAffected := query.RowsAffected; rowsAffected == 0 {
		transformer[quantityField] = 1
		if err := utils.DB.Table(ctrl.Table).Create(&transformer).Error; err != nil {
			ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
			return
		}
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "create "+ctrl.SingularLabel+" success", transformer))
}

func (ctrl CartController) Synch(ctx *gin.Context) {
	ctrl.Init(ctx)

	transformer, _ := utils.JsonFileParser(config.Data.SettingPath + "/transformers/request/" + ctrl.Table + "/synch.json")
	input := utils.ParseForm(ctx)

	if validation, err := utils.Validate(input, transformer); err {
		ctx.JSON(http.StatusOK, utils.ResponseData("failed", "validation", validation.Errors))
		return
	}

	utils.MapValuesShifter(transformer, input)
	utils.MapNullValuesRemover(transformer)

	setting := transformer["setting"].(map[string]any)
	userField := setting["user_id"].(string)
	quantityField := setting["quantity"].(string)

	delete(transformer, "setting")

	ids := (transformer["ids"]).([]any)
	quantities := transformer["quantities"].([]any)

	query := "UPDATE " + ctrl.Table + " SET " + quantityField + " = CASE "
	queryParams := []string{}

	for i, id := range ids {
		query += " WHEN id = " + strconv.Itoa(utils.ConvertToInt(id)) + " THEN " + strconv.Itoa(utils.ConvertToInt(quantities[i]))
		queryParams = append(queryParams, "id[]="+strconv.Itoa(utils.ConvertToInt(id)))
	}

	query += " ELSE `" + quantityField + "` END WHERE id IN (?) and `" + userField + "` = ?;"

	utils.DB.Exec(query, ids, transformer[userField])

	ctx.Redirect(http.StatusFound, "/api/v1/carts?"+strings.Join(queryParams, "&"))
	//ctx.JSON(http.StatusOK, utils.ResponseData("success", "update "+ctrl.SingularLabel+" success", transformer))
}

// todo : need to check constraint error
func (ctrl CartController) Delete(ctx *gin.Context) {
	ctrl.Init(ctx)

	transformer, _ := utils.JsonFileParser(config.Data.SettingPath + "/transformers/request/" + ctrl.Table + "/delete.json")

	setting := transformer["setting"].(map[string]any)
	userField := setting["user_id"].(string)

	delete(transformer, "setting")

	if err := utils.DB.Table(ctrl.Table).Where("id = ?", ctx.Param("id")).Where(userField+" = ?", ctx.Param("userId")).Delete(map[string]any{}).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, utils.ResponseData("error", err.Error(), nil))
		return
	}

	ctx.JSON(http.StatusOK, utils.ResponseData("success", "delete "+ctrl.SingularLabel+" success", nil))
}
