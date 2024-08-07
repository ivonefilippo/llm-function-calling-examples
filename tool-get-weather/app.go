package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/yomorun/yomo/serverless"
)

// Description outlines the functionality for the LLM Function Calling feature.
// It provides a detailed description of the function's purpose, essential for
// integration with LLM Function Calling. The presence of this function and its
// return value make the function discoverable and callable within the LLM
// ecosystem. For more information on Function Calling, refer to the OpenAI
// documentation at: https://platform.openai.com/docs/guides/function-calling
func Description() string {
	return `Get current weather for a given city. If no city is provided, you 
	should ask to clarify the city. If the city name is given, you should 
	convert the city name to Latitude and Longitude geo coordinates, keeping 
	Latitude and Longitude in decimal format.`
}

// InputSchema defines the argument structure for LLM Function Calling. It
// utilizes jsonschema tags to detail the definition. For jsonschema in Go,
// see https://github.com/invopop/jsonschema.
func InputSchema() any {
	return &LLMArguments{}
}

// LLMArguments defines the arguments for the LLM Function Calling. These
// arguments are combined to form a prompt automatically.
type LLMArguments struct {
	City      string  `json:"city" jsonschema:"description=The city name to get the weather for"`
	Latitude  float64 `json:"latitude" jsonschema:"description=The latitude of the city, in decimal format, range should be in (-90, 90)"`
	Longitude float64 `json:"longitude" jsonschema:"description=The longitude of the city, in decimal format, range should be in (-180, 180)"`
}

// Handler orchestrates the core processing logic of this function.
// - ctx.ReadLLMArguments() parses LLM Function Calling Arguments (skip if none).
// - ctx.WriteLLMResult() sends the retrieval result back to LLM.
// - ctx.Tag() identifies the tag of the incoming data.
// - ctx.Data() accesses the raw data.
// - ctx.Write() forwards processed data downstream.
func Handler(ctx serverless.Context) {
	var p LLMArguments
	// deserilize the arguments from llm tool_call response
	ctx.ReadLLMArguments(&p)

	// invoke the openweathermap api and return the result back to LLM
	result := requestOpenWeatherMapAPI(p.Latitude, p.Longitude)
	ctx.WriteLLMResult(result)

	slog.Info("get-weather", "city", p.City, "result", result)
}

func requestOpenWeatherMapAPI(lat, lon float64) string {
	const apiURL = "https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=metric"
	apiKey := os.Getenv("OPENWEATHERMAP_API_KEY")
	url := fmt.Sprintf(apiURL, lat, lon, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return "can not get the weather information at the moment"
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "can not get the weather information at the moment"
	}

	return string(body)
}

// DataTags specifies the data tags to which this serverless function
// subscribes, essential for data reception. Upon receiving data with these
// tags, the Handler function is triggered.
func DataTags() []uint32 {
	return []uint32{0x62}
}
