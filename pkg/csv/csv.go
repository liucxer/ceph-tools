package csv

import (
	"fmt"
	"reflect"
)

func ObjectToCsv(objectArgs interface{}) (string, string, error) {
	nameStr := ""
	valueStr := ""

	objectKind := reflect.TypeOf(objectArgs).Kind()

	objectRef := reflect.Value{}
	if objectKind == reflect.Struct {
		objectRef = reflect.ValueOf(objectArgs)
	} else if objectKind == reflect.Ptr {
		object := reflect.ValueOf(objectArgs)
		objectRef = object.Elem()
	}
	objectTypeList := objectRef.Type()
	for i := 0; i < objectRef.NumField(); i++ {
		objectField := objectRef.Field(i)
		name := objectTypeList.Field(i).Name
		if objectField.Kind() == reflect.Struct {
			subObject := objectField
			subObjectType := subObject.Type()
			for j := 0; j < subObject.NumField(); j++ {
				subObjectField := subObject.Field(j)
				nameStr += name + "." + subObjectType.Field(j).Name + ","
				valueStr += fmt.Sprintf("%v", subObjectField.Interface()) + ","
			}
		} else {
			nameStr += name + ","
			valueStr += fmt.Sprintf("%v", objectField.Interface()) + ","
		}
	}
	return nameStr, valueStr, nil
}

func ObjectListToCsv(object interface{}) (string, error) {
	sliceValue := reflect.ValueOf(object)
	sliceLen := sliceValue.Len()
	resNameStr := ""
	resValueStr := ""
	for i := 0; i < sliceLen; i++ {
		item := sliceValue.Index(i)

		nameStr := ""
		valueStr := ""

		objectTypeList := item.Type()
		for i := 0; i < item.NumField(); i++ {
			objectField := item.Field(i)
			name := objectTypeList.Field(i).Name
			if objectField.Kind() == reflect.Struct {
				subObject := objectField
				subObjectType := subObject.Type()
				for j := 0; j < subObject.NumField(); j++ {
					subObjectField := subObject.Field(j)
					nameStr += name + "." + subObjectType.Field(j).Name + ","
					valueStr += fmt.Sprintf("%v", subObjectField.Interface()) + ","
				}
			} else {
				nameStr += name + ","
				valueStr += fmt.Sprintf("%v", objectField.Interface()) + ","
			}
		}
		resNameStr = nameStr
		resValueStr += valueStr + "\n"
	}

	return resNameStr + "\n" + resValueStr, nil
}
