package utils
import (
	"github.com/gofiber/fiber/v2"
)

func GetIPAddress(c *fiber.Ctx) string {
	ip := c.IP()
	if ip == "::1" {
		ip = "127.0.0.1"
	}
	return ip
}