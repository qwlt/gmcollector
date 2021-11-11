package handlers

import (
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/qwlt/gmcollector/app/models"
	buff "github.com/qwlt/gmcollector/app/writebuffer"
)

var validate *validator.Validate

func TestHandler(c *fiber.Ctx) error {
	mv := MeasurementValidator{}
	if err := c.BodyParser(&mv); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{"errors": err.Error()})
	}

	err := validate.Struct(&mv)
	if err != nil {
		// log.Println(err.Error())
		responseMap := make(fiber.Map)
		responseMap["errors"] = make(fiber.Map)
		for _, err := range err.(validator.ValidationErrors) {
			key := err.Field()
			errorMessage := fmt.Sprintf("Validation error: %v", err.Tag())
			errors := responseMap["errors"].(fiber.Map)
			errors[key] = errorMessage
		}

		return c.Status(fiber.StatusBadRequest).JSON(&responseMap)
	}

	b, err := buff.GetBuffer()
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{"errors": err.Error()})
	}
	err = b.AddDatapoint(models.Measurement{DeviceID: mv.DeviceID, Value: mv.Value, Timestamp: mv.Timestamp})
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{"errors": err.Error()})
	}
	return c.SendStatus(fiber.StatusCreated)
}

func AnotherHandler(ctx *fiber.Ctx) error {

	return ctx.JSON(fiber.Map{
		"value": "OK",
	})
}

func InitValidator() {
	validate = validator.New()
}

func MainHandler(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{
		"message": "working",
	})
}
