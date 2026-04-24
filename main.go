package main

import "github.com/gofiber/fiber/v2"

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Type("html").SendString(`<!doctype html>
<html lang="es">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>go-media-api</title>
  <style>
    :root {
      color-scheme: light;
      --bg: #f4f7fb;
      --card: #ffffff;
      --text: #0f172a;
      --muted: #475569;
      --primary: #0ea5e9;
      --ok: #16a34a;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: "Segoe UI", system-ui, sans-serif;
      background: radial-gradient(circle at top left, #dbeafe, var(--bg));
      color: var(--text);
      min-height: 100vh;
      display: grid;
      place-items: center;
      padding: 1rem;
    }
    .card {
      width: min(680px, 100%);
      background: var(--card);
      border: 1px solid #e2e8f0;
      border-radius: 14px;
      padding: 1.25rem;
      box-shadow: 0 16px 40px rgba(15, 23, 42, 0.08);
    }
    h1 {
      margin: 0 0 0.25rem;
      font-size: clamp(1.4rem, 3vw, 1.9rem);
    }
    p {
      margin: 0.2rem 0;
      color: var(--muted);
    }
    code {
      background: #f1f5f9;
      border-radius: 8px;
      padding: 0.15rem 0.4rem;
      color: #0f172a;
    }
    .row {
      display: flex;
      flex-wrap: wrap;
      gap: 0.6rem;
      margin-top: 1rem;
      align-items: center;
    }
    a.button {
      text-decoration: none;
      color: #fff;
      background: var(--primary);
      padding: 0.55rem 0.9rem;
      border-radius: 10px;
      font-weight: 600;
    }
    .status {
      color: var(--ok);
      font-weight: 700;
    }
  </style>
</head>
<body>
  <main class="card">
    <h1>go-media-api</h1>
    <p>Interfaz minima del binario.</p>
    <p>Puerto por defecto: <code>:8080</code></p>
    <p>Health actual: <span class="status">ok</span></p>
    <div class="row">
      <a class="button" href="/health">Ver /health</a>
      <a class="button" href="/info">Ver /info</a>
    </div>
  </main>
</body>
</html>`)
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})

	app.Get("/info", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"name":    "go-media-api",
			"version": "dev",
			"port":    "8080",
		})
	})

	if err := app.Listen(":8080"); err != nil {
		panic(err)
	}
}
