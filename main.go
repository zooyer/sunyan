package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"unicode"

	"github.com/Luxurioust/excelize"
	"github.com/zooyer/jsons"
)

func toFloat64(str string) (float64, error) {
	var n float64
	if err := json.Unmarshal([]byte(str), &n); err != nil {
		return 0, err
	}
	return n, nil
}

func toInt64(str string) (int64, error) {
	var n int64
	if err := json.Unmarshal([]byte(str), &n); err != nil {
		return 0, err
	}
	return n, nil
}

func toNumber(str string) (json.Number, error) {
	var n json.Number
	if err := json.Unmarshal([]byte(str), &n); err != nil {
		return "", err
	}
	return n, nil
}

func isNumber(str string) bool {
	_, err := toNumber(str)
	return err == nil
}

func isSpace(str string) bool {
	for _, c := range str {
		if !unicode.IsSpace(c) {
			return false
		}
	}
	return true
}

func allIsNil(s []string) bool {
	for _, s := range s {
		if !isSpace(s) {
			return false
		}
	}

	return true
}

// 解析首行
func unmarshalRowFirst(row []string) jsons.Object {
	var slice []string
	var obj = make(jsons.Object)

	for _, r := range row {
		if !isSpace(r) {
			slice = append(slice, r)
		}
	}

	if len(slice) > 1 {
		obj[slice[0]] = slice[1]
	}

	if len(slice) > 2 {
		obj["日期"] = slice[2]
	}

	return obj
}

// 解析操作人员信息
func unmarshalOperator(rows [][]string) jsons.Object {
	var opt = make(jsons.Object)

	var slice []string
	for i := len(rows) - 1; i >= 0; i-- {
		if !allIsNil(rows[i]) {
			for _, r := range rows[i] {
				if !isSpace(r) {
					slice = append(slice, r)
				}
			}
			break
		}
	}

	if len(slice) > 1 {
		opt[slice[0]] = slice[1]
		slice = slice[2:]
	}

	if len(slice) > 1 {
		opt[slice[len(slice)-2]] = slice[len(slice)-1]
		slice = slice[:len(slice)-2]
	}

	if len(slice) > 1 {
		opt[slice[0]] = slice[1:]
	}

	return opt
}

// 解析纵向表单
/*
  ----------------------------------------
  - name  - key1 - key2 - key3 - key4
  ----------------------------------------
  - name1 - val1 - val2 - val3 - val4
  ----------------------------------------
  - name2 - val1 - val2 - val3 - val4
  ----------------------------------------
*/
func unmarshalTableVertical(rows [][]string, x, y, width, height int) (string, jsons.Object) {
	var obj = make(jsons.Object)
	for i := 1; i < height; i++ {
		if rows[i+y][x] == "" {
			continue
		}
		obj[rows[i+y][x]] = make(jsons.Object)
		for j := 1; j < width; j++ {
			key := rows[y][j+x]
			val := rows[i+y][j+x]
			if key != "" || val != "" {
				obj.Get(rows[i+y][x]).Object()[key] = val
			}
		}
	}

	return rows[y][x], obj
}

// 解析横向表单
/*
  ----------------------------------------
  - name - name1 - name2 - name3 - name4 -
  ----------------------------------------
  - key1 - val1  - val2  - val3  - val4  -
  ----------------------------------------
  - key2 - val1  - val2  - val3  - val4  -
  ----------------------------------------
*/
func unmarshalTableHorizontal(rows [][]string, x, y, width, height int) (string, jsons.Object) {
	var obj = make(jsons.Object)
	for i := 1; i < height; i++ {
		for j := 1; j < width; j++ {
			if len(rows[y+i]) <= j {
				continue
			}
			if rows[y][x+j] == "" {
				continue
			}
			if obj[rows[y][x+j]] == nil {
				obj[rows[y][x+j]] = make(jsons.Object)
			}
			key := rows[y+i][x]
			val := rows[y+i][j+x]
			if key != "" || val != "" {
				obj.Get(rows[y][x+j]).Object()[key] = val
			}
		}
	}

	return rows[y][x], obj
}

// 解析高度为2的键值对
/* 宽度为5的kv键值对
------------------------------------
- key1 - key2 - key3 - key4 - key5 -
------------------------------------
- val1 - val2 - val3 - val4 - val5 -
------------------------------------
*/
func unmarshalHorizontal(rows [][]string, x, y, width int) jsons.Object {
	var obj = make(jsons.Object)

	for i := 0; i < width; i++ {
		key := rows[y][x+i]
		val := rows[y+1][x+i]
		if key != "" || val != "" {
			obj[key] = val
		}
	}

	return obj
}

// 解析宽度为2的键值对
/* 高度为5的kv键值对
---------------
- key1 - val1 -
---------------
- key2 - val2 -
---------------
- key3 - val3 -
---------------
- key4 - val4 -
---------------
- key5 - val5 -
---------------
*/
func unmarshalVertical(rows [][]string, x, y, height int) jsons.Object {
	var obj = make(jsons.Object)

	for i := 0; i < height; i++ {
		key := rows[y+i][x]
		val := rows[y+i][x+1]
		if key != "" || val != "" {
			obj[key] = val
		}
	}

	return obj
}

// 解析宽度为n的键值对，过滤掉中间的空白字段
/*
------------------
- key1 -  - val1 -
------------------
- key2 -  - val2 -
------------------
- key3 -  - val3 -
------------------
- key4 -  - val4 -
------------------
- key5 -  - val5 -
------------------
 */
func unmarshalMultiVertical(rows [][]string, x, y, width, height int) jsons.Object {
	var obj = make(jsons.Object)

	for i := 0; i < height; i++ {
		if row := rows[y+i][x : x+width]; len(row) > 1 {
			var slice []string
			for _, r := range row {
				if !isSpace(r) {
					slice = append(slice, r)
				}
			}
			if len(slice) > 1 {
				obj[slice[0]] = slice[1]
			}
		}
	}

	return obj
}

func main() {
	dirs, err := ioutil.ReadDir("./")
	if err != nil {
		panic(err)
	}

	// 遍历当前目录下所有文件
	for _, dir := range dirs {
		if dir.IsDir() {
			continue
		}
		if !strings.HasSuffix(dir.Name(), ".xlsx") {
			continue
		}

		// 文件名
		name := strings.TrimSuffix(dir.Name(), ".xlsx")

		xlsx, err := excelize.OpenFile(dir.Name())
		if err != nil {
			panic(err)
		}

		var doc = make(jsons.Object)

		// 遍历sheet
		for i := 0; i < xlsx.SheetCount; i++ {
			// 获取sheet名称
			name := xlsx.GetSheetName(i)

			// 获取rows
			rows, err := xlsx.GetRows(name)
			if err != nil {
				panic(err)
			}

			if isNumber(name) {
				var page = make(jsons.Object)

				// 首行基础信息
				page["基础信息"] = unmarshalRowFirst(rows[0])

				// 操作人员
				page["操作人员"] = unmarshalOperator(rows)

				// 炉号工艺卡
				page["炉号工艺卡"] = unmarshalHorizontal(rows, 0, 1, 26)

				// 成分表
				_, table := unmarshalTableHorizontal(rows, 0, 3, 26, 10)
				page["成分表"] = table

				// 操作时间
				_, table = unmarshalTableVertical(rows, 0, 14, 4, 16)
				page["操作时间"] = table

				// 治炼时间
				_, table = unmarshalTableVertical(rows, 4, 14, 3, 16)
				page["治炼时间"] = table

				// 操作造渣参数
				_, table = unmarshalTableVertical(rows, 7, 14, 3, 5)
				page["操作造渣参数"] = table

				// 治炼造渣参数
				_, table = unmarshalTableVertical(rows, 10, 14, 4, 5)
				page["治炼造渣参数"] = table

				// 操作温度参数
				_, table = unmarshalTableVertical(rows, 7, 20, 3, 5)
				page["操作温度参数"] = table

				// 治炼温度参数
				_, table = unmarshalTableVertical(rows, 10, 20, 4, 5)
				page["治炼温度参数"] = table

				// 电能消耗
				page["电能消耗"] = unmarshalMultiVertical(rows, 7, 26, 7, 4)

				// 成分调整时间
				_, table = unmarshalTableVertical(rows, 14, 14, 6, 16)
				page["成分调整时间"] = table

				// 渣料统计
				_, table = unmarshalTableVertical(rows, 20, 14, 5, 8)
				page["渣料统计"] = table

				// 治炼结果
				_, table = unmarshalTableHorizontal(rows, 0, 30, 10, 8)
				page["治炼结果"] = table

				doc[name] = page
			}
		}

		data, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			panic(err)
		}

		if err = ioutil.WriteFile(fmt.Sprintf("%s.json", name), data, 0644); err != nil {
			panic(err)
		}
	}
}
