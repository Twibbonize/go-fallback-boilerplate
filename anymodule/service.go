package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"

	moduleboilerplate "github.com/Twibbonize/go-module-boilerplate-mongodb"
)

type server struct {
	anyModuleSetter moduleboilerplate.SetterLib
	anyModuleGetter moduleboilerplate.GetterLib
}


func (s *server) SeedOneByRandId(ctx *fiber.Ctx) error {

	type Body struct {
		RandId string `json:"randid" binding:"required"`
	}

	var requestBody Body
	if errorBind := ctx.BodyParser(&requestBody); errorBind != nil {
		return ConstructResponse(ctx, http.StatusBadRequest, "RandId is required")
	}

	_, err := s.anyModuleGetter.GetByRandID(requestBody.RandId)
	if err != nil {
		return ConstructResponse(ctx, fiber.StatusNotFound, "Module not found")
	}
	return ConstructResponse(ctx, fiber.StatusOK, "Module fetched successfully")
}

func (s *server) SeedOneByUUID(ctx *fiber.Ctx) error {

	type Body struct {
		Uuid string `json:"uuid" binding:"required"`
	}

	var requestBody Body
	if errorBind := ctx.BodyParser(&requestBody); errorBind != nil {
		return ConstructResponse(ctx, http.StatusBadRequest, "uuid is required")
	}

	_, err := s.anyModuleSetter.FindByUUID(requestBody.Uuid, true)

	if err != nil {
		return ConstructResponse(ctx, fiber.StatusNotFound, "Module not found",)
	}

	return ConstructResponse(ctx, fiber.StatusOK, "Module fetched successfully") // No data is returned in this method
}

func (s *server) SeedMany(ctx *fiber.Ctx) error {

	type Body struct {
		RetrievedLengthStr string `json:"retrievedlengthstr" binding:"required"`
		LastObjectIdHex string `json:"lastobjectidhex" binding:"required"`
		ValidLastUUID string `json:"validlastuuid" binding:"required"`
		CampaignUUID string `json:"campaignuuid" binding:"required"`
	}

	var requestBody Body
	if errorBind := ctx.BodyParser(&requestBody); errorBind != nil {
		return ConstructResponse(ctx, http.StatusBadRequest, "retrievedlengthstr, lastobjectidhex, validlastuuid, campaignuuid is required")
	}

	if os.Getenv("APP_ENV") == "development" {
		fmt.Println("GetSetModule")
		fmt.Println("requestBody", requestBody)
	}

	// Parse retrievedLengthStr as int
	retrievedLength, err := strconv.ParseInt(requestBody.RetrievedLengthStr, 10, 64)
	if err != nil {
		 return ConstructResponse(ctx, fiber.StatusBadRequest, "Invalid retrievedLength value")
	}

	errFind := s.anyModuleSetter.SeedLinked(retrievedLength, requestBody.LastObjectIdHex, requestBody.ValidLastUUID, requestBody.CampaignUUID)
	if errFind != nil {
		 return ConstructResponse(ctx, fiber.StatusInternalServerError, errFind.Err.Error())
	}

	return ConstructResponse(ctx, fiber.StatusOK, "Module fetched successfully")
}

func (s *server) DeleteManyByParticipant(ctx *fiber.Ctx) error {

	type Body struct {
		Uuid string `json:"uuid" binding:"required"`
	}

	var requestBody Body
	if errorBind := ctx.BodyParser(&requestBody); errorBind != nil {
		return ConstructResponse(ctx, http.StatusBadRequest, "uuid is required")
	}

	err := s.anyModuleSetter.DeleteManyByAnyUUID(requestBody.Uuid)
	if err != nil {
		 return ConstructResponse(ctx, fiber.StatusInternalServerError, "Failed to delete submissions")
	}

	return ConstructResponse(ctx, fiber.StatusOK, "Module deleted successfully")
}