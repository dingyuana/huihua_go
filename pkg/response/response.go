package response

import "github.com/gofiber/fiber/v2"

// R is a unified response helper.
var R = &Response{}

// Response provides unified response methods.
type Response struct{}

// Ok returns a 200 JSON response with the given data.
func (r *Response) Ok(c *fiber.Ctx, data interface{}) error {
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

// Created returns a 201 JSON response with the given data.
func (r *Response) Created(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"code":    0,
		"message": "created",
		"data":    data,
	})
}

// OkMessage returns a 200 JSON response with only a message.
func (r *Response) OkMessage(c *fiber.Ctx, message string) error {
	return c.JSON(fiber.Map{
		"code":    0,
		"message": message,
	})
}

// Error returns an error JSON response with the given status code and message.
func (r *Response) Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"code":    status,
		"message": message,
	})
}

// BadRequest returns a 400 error response.
func (r *Response) BadRequest(c *fiber.Ctx, message string) error {
	return r.Error(c, fiber.StatusBadRequest, message)
}

// NotFound returns a 404 error response.
func (r *Response) NotFound(c *fiber.Ctx, message string) error {
	return r.Error(c, fiber.StatusNotFound, message)
}

// InternalError returns a 500 error response.
func (r *Response) InternalError(c *fiber.Ctx, message string) error {
	return r.Error(c, fiber.StatusInternalServerError, message)
}

// List returns a paginated list response.
func (r *Response) List(c *fiber.Ctx, items interface{}, total int) error {
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "success",
		"data":    items,
		"total":   total,
	})
}
