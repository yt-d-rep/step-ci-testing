package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"

	"step-ci-testing/oapi"
)

// ServerInterfaceを満たすよう実装
type controller struct {}

func (c controller) GetMorphs(ctx echo.Context) error {
	morphs := readData("data.json")
	return ctx.JSON(http.StatusOK, &morphs)
}

func (c controller) PostMorph(ctx echo.Context) error {
	newMorph := new(oapi.NewMorph)
	ctx.Bind(newMorph)
	if len(newMorph.Name) > 64 {
		resp := &oapi.BadRequest{"Invalid value"}
		return ctx.JSON(http.StatusBadRequest, resp)
	}
	morph := writeData("data.json", newMorph)
	return ctx.JSON(http.StatusOK, &morph)
}

func (c controller) GetMorphById(ctx echo.Context, id oapi.Id) error {
	morphs := readData("data.json")
	for _, m := range morphs {
		if m.Id == id {
			return ctx.JSON(http.StatusOK, &m)
		}
	}
	resp := &oapi.NotFound{"Not found"}
	return ctx.JSON(http.StatusNotFound, &resp)
}

// dataaccess
func readData(filepath string) oapi.Morphs {
	data, _ := ioutil.ReadFile(filepath)
	var morphs oapi.Morphs
	json.Unmarshal(data, &morphs)
	return morphs
}

func writeData(filepath string, newMorph *oapi.NewMorph) oapi.Morph {
	morphs := readData(filepath)

	id := len(morphs) + 1
	morph := oapi.Morph{
		Id: id,
		Name: newMorph.Name,
	}
	morphs = append(morphs, morph)

	f, err := os.OpenFile(filepath, os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(morphs); err != nil {
		panic(err)
	}

	return morph
}

func main() {
	e := echo.New()
	c := controller{}

	oapi.RegisterHandlers(e, c)

	e.Logger.Fatal(e.Start(":8888"))
}
