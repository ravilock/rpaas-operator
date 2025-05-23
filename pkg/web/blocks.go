// Copyright 2019 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/tsuru/rpaas-operator/internal/pkg/rpaas"
	"github.com/tsuru/rpaas-operator/pkg/macro"
)

func deleteBlock(c echo.Context) error {
	ctx := c.Request().Context()
	manager, err := getManager(ctx)
	if err != nil {
		return err
	}

	serverName := c.QueryParam("server_name")

	err = manager.DeleteBlock(ctx, c.Param("instance"), serverName, c.Param("block"))
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func listBlocks(c echo.Context) error {
	ctx := c.Request().Context()
	manager, err := getManager(ctx)
	if err != nil {
		return err
	}

	blocks, err := manager.ListBlocks(ctx, c.Param("instance"))
	if err != nil {
		return err
	}

	if blocks == nil {
		blocks = make([]rpaas.ConfigurationBlock, 0)
	}

	return c.JSON(http.StatusOK, struct {
		Blocks []rpaas.ConfigurationBlock `json:"blocks"`
	}{blocks})
}

func updateBlock(c echo.Context) error {
	if c.Request().ContentLength == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Request body can't be empty")
	}
	ctx := c.Request().Context()
	manager, err := getManager(ctx)
	if err != nil {
		return err
	}

	var block rpaas.ConfigurationBlock
	if err = c.Bind(&block); err != nil {
		return err
	}

	_, err = macro.ExpandWithOptions(block.Content, macro.ExpandOptions{IgnoreSyntaxErrors: false})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = manager.UpdateBlock(ctx, c.Param("instance"), block)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func deleteLuaBlock(c echo.Context) error {
	ctx := c.Request().Context()
	manager, err := getManager(ctx)
	if err != nil {
		return err
	}

	luaBlockType, err := formValue(c.Request(), "lua_module_type")
	if err != nil {
		return err
	}

	err = manager.DeleteBlock(ctx, c.Param("instance"), "", luaBlockName(luaBlockType))
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func listLuaBlocks(c echo.Context) error {
	ctx := c.Request().Context()
	manager, err := getManager(ctx)
	if err != nil {
		return err
	}

	blocks, err := manager.ListBlocks(ctx, c.Param("instance"))
	if err != nil {
		return err
	}

	type luaBlock struct {
		LuaName string `json:"lua_name"`
		Content string `json:"content"`
	}

	luaBlocks := []luaBlock{}
	for _, block := range blocks {
		if strings.HasPrefix(block.Name, luaBlockName("")) {
			luaBlocks = append(luaBlocks, luaBlock{
				LuaName: block.Name,
				Content: block.Content,
			})
		}
	}

	return c.JSON(http.StatusOK, struct {
		Modules []luaBlock `json:"modules"`
	}{Modules: luaBlocks})
}

func updateLuaBlock(c echo.Context) error {
	ctx := c.Request().Context()
	manager, err := getManager(ctx)
	if err != nil {
		return err
	}

	in := struct {
		LuaModuleType string `form:"lua_module_type"`
		Content       string `form:"content"`
	}{}
	if err = c.Bind(&in); err != nil {
		return err
	}

	block := rpaas.ConfigurationBlock{
		Name:    luaBlockName(in.LuaModuleType),
		Content: in.Content,
	}

	err = manager.UpdateBlock(ctx, c.Param("instance"), block)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func luaBlockName(blockType string) string {
	return fmt.Sprintf("lua-%s", blockType)
}
