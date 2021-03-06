/**
 * Initialize base CRUD controller class.  Malina eCommerce application
 *
 *
 * @author		John Aran (Ilyas Toxanbayev)
 * @version		1.0.0
 * @based on
 * @email      		il.aranov@gmail.com
 * @link
 * @github      	https://github.com/ilyaran/Malina
 * @license		MIT License Copyright (c) 2017 John Aran (Ilyas Toxanbayev)
 */

package controller

import (
	"net/http"
	"regexp"
	"Malina/config"
	"strings"
	"Malina/libraries"
	"strconv"
	"github.com/gorilla/mux"
	"fmt"
	"Malina/models"
)

/**
* Errors Code
* error code 0<=  x < 100. Server Error Input Errors (MaxBytesReader, Parse Post Form, bad request)
* error code 100<=  x < 200. Validation Errors
* error code 200<=  Database Errors
*
* 10 - ["error"] = "bad request"  Post Get Put
* 30 - ["error"] = Server error  Post Get Put
*
* 100 [keyName] = error of validation example "Item id must an integer"
* 110 ["error"] = not find
* 120 ["error"] = exceeds login attempts
* 130 ["error"] = common validation error
* 140 ["error"] = other
*
* 200 NOT exists(SELECT category_id
* 201 Parent Id is exists
* 210 images NOT exists
* 220 NOT exists(SELECT account_id
* 230 NOT exists(SELECT item_id
* 250 result no rows

**/


type CrudController struct {
	dbtable  string
	order_by string
}

func (s *CrudController) auth(redirect string, w http.ResponseWriter, r *http.Request) {
	library.SESSION.Authentication(w, r)
	if library.SESSION.GetSessionObj().GetAccount_id() == 0 {
		if redirect != "" {
			http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
			return
		}
	}
	library.VALIDATION.Status = 0
	library.VALIDATION.Result = map[string]string{}
}
func (s *CrudController) authAdmin(dbtable string, w http.ResponseWriter, r *http.Request) string {
	s.dbtable = dbtable
	vars := mux.Vars(r)
	action := vars["action"]
	library.SESSION.Authentication(w, r)
	if library.SESSION.GetSessionObj().GetPermission() != "admin" {
		if action == "list" {
			//w.Write([]byte(errors.ErrorView("Not permission","access denied")))
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return ""
		}
		library.VALIDATION.Status = 80
		library.VALIDATION.Result = map[string]string{"unauth":"access denied"}
		return ""
	}
	library.VALIDATION.Status = 0
	library.VALIDATION.Result = map[string]string{}
	return action
}
func (s *CrudController) getList(enable bool, w http.ResponseWriter, r *http.Request) (int64, string) {
	return library.VALIDATION.IsInt64(false, "page", 20, r)
}
func (s *CrudController) getAjaxList(enable bool, w http.ResponseWriter, r *http.Request) (int64, string, int64, string, string, int64) {
	page, pageStr := library.VALIDATION.IsInt64(false, "page", 20, r)
	per_page, per_pageStr := library.VALIDATION.IsInt64(false, "per_page", 4, r)
	if per_page < app.Per_page() {
		per_page = app.Per_page()
		per_pageStr = strconv.FormatInt(app.Per_page(), 10)
	}
	order_by, _ := library.VALIDATION.IsInt64(false, "order_by", 2, r)
	var search = r.FormValue("search")
	if search != "" {
		if len(search) > 64 {
			search = search[0:64]
		}
		if s.dbtable != "product" {
			search = regexp.MustCompile(`[\W]`).ReplaceAllString(search, "")
		} else {
			search = strings.Trim(regexp.MustCompile(`[\W]+`).ReplaceAllString(search, "|"), "|")
		}
		//fmt.Println(search)
	}

	return page, pageStr, per_page, per_pageStr, search, order_by
}

func (s *CrudController) get(w http.ResponseWriter, r *http.Request) (int64, string) {
	idInt64, idStr := library.VALIDATION.IsInt64(true, s.dbtable + "_id", 20, r)
	return idInt64, idStr
}
func (s *CrudController) add(w http.ResponseWriter, r *http.Request) {

}
func (s *CrudController) edit(w http.ResponseWriter, r *http.Request) (int64, string) {
	idInt64, idStr := library.VALIDATION.IsInt64(true, s.dbtable + "_id", 20, r)
	return idInt64, idStr
}
func (s *CrudController) del(w http.ResponseWriter, r *http.Request) (int64, string) {
	idInt64, idStr := library.VALIDATION.IsInt64(true, s.dbtable + "_id", 20, r)
	return idInt64, idStr
}

func (s *CrudController) inlist(columns map[string]string, r *http.Request) {
	r.ParseForm()
	for columnName, columnType := range columns {
		s.inlistUpdateQuery(columnName, columnType, r)
	}
	s.inlistDeleteQuery(r)
}

func (s *CrudController) inlistUpdateQuery(columnName, columnType string, r *http.Request) {
	var valueNum int = 1 // count numbers 1,2,3 ... to "$1,$2,$3,$4,$5 ..." string
	var exec []interface{}
	var layout_item_query = `UPDATE ` + s.dbtable + `
		SET ` + columnName + ` = data_table.v
		FROM (
		       SELECT
			 unnest(ARRAY[ %s ]) AS v,
			 unnest(ARRAY[ %s ]) AS i
		     )AS data_table
		WHERE ` + s.dbtable + `_id = data_table.i
		`
	var values string
	var ids string
	var match bool
	for _, v1 := range r.PostForm[columnName + "_inlist[]"] {
		// The string - 25|89  or  25|253.56  or  25|Some Title
		// is splitting to array [25,89] or [25,253.56]  or  [25,Some Title]
		id_value_arr := strings.SplitN(v1, "|", 2)
		// id_value_arr[0] == 25
		// id_value_arr[1] == 89 or 253.56 or Some Title
		match, _ = regexp.MatchString(`^[0-9]{1,20}$`, id_value_arr[0])
		if match {
			// push value to exec interface array
			switch columnType {
			case "integer":
				match, _ = regexp.MatchString(`^[0-9]+$`, id_value_arr[1])
				if match {
					values += `,` + id_value_arr[1]
					ids += `,` + id_value_arr[0]
					valueNum++
				}
			case "numeric":
				match, _ = regexp.MatchString(`^[0-9]{1,20}(\.[0-9]{0,2})?$`, id_value_arr[1])
				if match {
					values += `,` + id_value_arr[1]
					ids += `,` + id_value_arr[0]
					valueNum++
				}

			case "boolean":
				// building "TRUE,TRUE,TRUE,FALSE, ..." string to values array: unnest(ARRAY[TRUE,TRUE,TRUE,FALSE, ...]) AS v
				valueBool, _ := strconv.ParseBool(id_value_arr[1])
				if valueBool {
					values += `,TRUE`
				} else {
					values += `,FALSE`
				}
				ids += `,` + id_value_arr[0]
				valueNum++
			case "string":
				match, _ = regexp.MatchString(app.Pattern_title(), id_value_arr[1])
				if match {
					// building "$1,$2,$3,$4,$5 ..." string to values array: unnest(ARRAY[$1,$2,$3,$4,$5 ...]) AS v
					values += `,$` + strconv.Itoa(valueNum)
					exec = append(exec, id_value_arr[1])
					ids += `,` + id_value_arr[0]
					valueNum++
				}
			}
		}
	}
	if ids != `` {
		//fmt.Println(fmt.Sprintf(layout_item_query, values[1:], ids[1:]))
		if model.Crud.Update(fmt.Sprintf(layout_item_query, values[1:], ids[1:]), exec) > 0 {
			return
		} else {
			library.VALIDATION.Status = 200
			library.VALIDATION.Result[columnName] = "not updated"
			return
		}
	}
}
func (s *CrudController) inlistDeleteQuery(r *http.Request) {
	// delete in list
	var del_ids string
	for _, v2 := range r.PostForm["del_inlist[]"] {
		values_arr := strings.SplitN(v2, "|", 2)
		match, _ := regexp.MatchString(`^[0-9]{1,20}$`, values_arr[0])
		if match {
			valueBool, _ := strconv.ParseBool(values_arr[1])
			if valueBool {
				del_ids += `,` + values_arr[0]
			}
		}
	}
	var del_q string
	if del_ids != `` {
		del_q = fmt.Sprintf(`
		DELETE FROM ` + s.dbtable + ` WHERE ` + s.dbtable + `_id = ANY(ARRAY[%s])`, del_ids[1:])
		if model.Crud.Delete(del_q, []interface{}{}) > 0 {
			return
		} else {
			library.VALIDATION.Status = 200
			library.VALIDATION.Result[s.dbtable + "_delete"] = "not deleted"
			return
		}
	}
}
