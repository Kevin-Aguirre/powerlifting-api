package api

import (
	"encoding/json"
	"net/http"
)

func openAPIHandler() http.HandlerFunc {
	spec := buildOpenAPISpec()
	data, _ := json.Marshal(spec)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

func buildOpenAPISpec() map[string]interface{} {
	str := func(s string) map[string]interface{} { return map[string]interface{}{"type": "string"} }
	num := func() map[string]interface{} { return map[string]interface{}{"type": "number"} }
	integer := func() map[string]interface{} { return map[string]interface{}{"type": "integer"} }
	_ = str

	paginationParams := []map[string]interface{}{
		{
			"name": "limit", "in": "query", "schema": map[string]interface{}{"type": "integer", "default": 50, "maximum": 200},
			"description": "Results per page",
		},
		{
			"name": "offset", "in": "query", "schema": integer(),
			"description": "Number of results to skip",
		},
	}

	unitParam := map[string]interface{}{
		"name": "unit", "in": "query", "schema": map[string]interface{}{"type": "string", "enum": []string{"kg", "lbs"}},
		"description": "Weight unit (default: kg)",
	}

	sortParam := func(fields []string, def string) map[string]interface{} {
		return map[string]interface{}{
			"name": "sort", "in": "query",
			"schema": map[string]interface{}{"type": "string", "enum": fields, "default": def},
			"description": "Field to sort by",
		}
	}

	orderParam := map[string]interface{}{
		"name": "order", "in": "query",
		"schema": map[string]interface{}{"type": "string", "enum": []string{"asc", "desc"}, "default": "asc"},
		"description": "Sort order",
	}

	paginatedResponse := func(dataSchema map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"total":  integer(),
				"limit":  integer(),
				"offset": integer(),
				"data":   dataSchema,
			},
		}
	}

	liftAttemptsSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"attempt1": num(), "attempt2": num(), "attempt3": num(), "attempt4": num(), "best": num(),
		},
	}

	meetResultSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"place": map[string]interface{}{"type": "string"},
			"name":  map[string]interface{}{"type": "string"},
			"sex":   map[string]interface{}{"type": "string"},
			"age":   num(),
			"country": map[string]interface{}{"type": "string"},
			"equipment": map[string]interface{}{"type": "string"},
			"division":  map[string]interface{}{"type": "string"},
			"bodyweightKg":  num(),
			"weightClassKg": map[string]interface{}{"type": "string"},
			"squat":    liftAttemptsSchema,
			"bench":    liftAttemptsSchema,
			"deadlift": liftAttemptsSchema,
			"totalKg":  num(),
			"event":    map[string]interface{}{"type": "string"},
			"tested":   map[string]interface{}{"type": "string"},
			"meetDate":       map[string]interface{}{"type": "string"},
			"meetFederation": map[string]interface{}{"type": "string"},
			"meetName":       map[string]interface{}{"type": "string"},
		},
	}

	pbSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"squat": num(), "bench": num(), "deadlift": num(), "total": num(), "dots": num(),
		},
	}

	lifterSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{"type": "string"},
			"pb": map[string]interface{}{
				"type":                 "object",
				"additionalProperties": pbSchema,
				"description":          "Keyed by equipment type",
			},
			"competitionResults": map[string]interface{}{
				"type":  "array",
				"items": meetResultSchema,
			},
		},
	}

	meetSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"federation":  map[string]interface{}{"type": "string"},
			"date":        map[string]interface{}{"type": "string"},
			"meetCountry": map[string]interface{}{"type": "string"},
			"meetState":   map[string]interface{}{"type": "string"},
			"meetTown":    map[string]interface{}{"type": "string"},
			"meetName":    map[string]interface{}{"type": "string"},
			"ruleSet":     map[string]interface{}{"type": "string"},
		},
	}

	recordHolderSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"lift": num(), "lifter": map[string]interface{}{"type": "string"},
		},
	}

	recordSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"weightClassKg": map[string]interface{}{"type": "string"},
			"sex":           map[string]interface{}{"type": "string"},
			"equipment":     map[string]interface{}{"type": "string"},
			"squat":         recordHolderSchema,
			"bench":         recordHolderSchema,
			"deadlift":      recordHolderSchema,
			"total":         recordHolderSchema,
		},
	}

	prEntrySchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"date": map[string]interface{}{"type": "string"},
			"squat": num(), "bench": num(), "deadlift": num(), "total": num(), "dots": num(),
		},
	}

	careerStatsSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name":              map[string]interface{}{"type": "string"},
			"totalCompetitions": integer(),
			"firstCompetition":  map[string]interface{}{"type": "string"},
			"lastCompetition":   map[string]interface{}{"type": "string"},
			"federations":       map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
			"prProgression": map[string]interface{}{
				"type":                 "object",
				"additionalProperties": map[string]interface{}{"type": "array", "items": prEntrySchema},
				"description":          "Keyed by equipment type",
			},
		},
	}

	sexEquipWcParams := []map[string]interface{}{
		{"name": "sex", "in": "query", "schema": map[string]interface{}{"type": "string", "enum": []string{"M", "F", "Mx"}}, "description": "Filter by sex"},
		{"name": "equipment", "in": "query", "schema": map[string]interface{}{"type": "string"}, "description": "Filter by equipment (e.g. Raw, Single-ply)"},
		{"name": "weightClass", "in": "query", "schema": map[string]interface{}{"type": "string"}, "description": "Filter by weight class kg (e.g. 83)"},
	}

	meetFilterParams := []map[string]interface{}{
		{"name": "country", "in": "query", "schema": map[string]interface{}{"type": "string"}, "description": "Filter by country"},
		{"name": "from", "in": "query", "schema": map[string]interface{}{"type": "string", "format": "date"}, "description": "Filter meets on or after this date (YYYY-MM-DD)"},
		{"name": "to", "in": "query", "schema": map[string]interface{}{"type": "string", "format": "date"}, "description": "Filter meets on or before this date (YYYY-MM-DD)"},
	}

	jsonResponse := func(schema map[string]interface{}) map[string]interface{} {
		return map[string]interface{}{
			"200": map[string]interface{}{
				"description": "OK",
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{"schema": schema},
				},
			},
		}
	}

	concatParams := func(slices ...[]map[string]interface{}) []map[string]interface{} {
		var out []map[string]interface{}
		for _, s := range slices {
			out = append(out, s...)
		}
		return out
	}

	appendParam := func(params []map[string]interface{}, p ...map[string]interface{}) []map[string]interface{} {
		return append(params, p...)
	}

	return map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":       "Powerlifting API",
			"version":     "1.0.0",
			"description": "REST API for OpenPowerlifting data (~1M lifters, ~62K meets, ~250 federations).",
		},
		"paths": map[string]interface{}{
			"/health": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Health check",
					"responses": jsonResponse(map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"status":      map[string]interface{}{"type": "string"},
							"lastUpdated": map[string]interface{}{"type": "string", "format": "date-time"},
						},
					}),
				},
			},
			"/lifters": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "List all lifters",
					"parameters": concatParams(paginationParams, []map[string]interface{}{unitParam, orderParam,
						sortParam([]string{"name"}, "name"),
						{"name": "sex", "in": "query", "schema": map[string]interface{}{"type": "string"}, "description": "Filter by sex (M/F/Mx)"},
						{"name": "equipment", "in": "query", "schema": map[string]interface{}{"type": "string"}, "description": "Filter by equipment type"},
					}),
					"responses": jsonResponse(paginatedResponse(map[string]interface{}{"type": "array", "items": lifterSchema})),
				},
			},
			"/lifters/names": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "List all lifter names",
					"parameters": concatParams(paginationParams, []map[string]interface{}{orderParam,
						sortParam([]string{"name"}, "name"),
						{"name": "sex", "in": "query", "schema": map[string]interface{}{"type": "string"}, "description": "Filter by sex"},
						{"name": "equipment", "in": "query", "schema": map[string]interface{}{"type": "string"}, "description": "Filter by equipment"},
					}),
					"responses": jsonResponse(paginatedResponse(map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}})),
				},
			},
			"/lifters/search": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Search lifters by partial name",
					"parameters": concatParams(paginationParams, []map[string]interface{}{unitParam, orderParam,
						sortParam([]string{"name"}, "name"),
						{"name": "q", "in": "query", "required": true, "schema": map[string]interface{}{"type": "string"}, "description": "Partial name query"},
					}),
					"responses": jsonResponse(paginatedResponse(map[string]interface{}{"type": "array", "items": lifterSchema})),
				},
			},
			"/lifters/top": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Top lifters ranked by DOTS score",
					"parameters": concatParams(paginationParams, sexEquipWcParams, []map[string]interface{}{unitParam}),
					"responses": jsonResponse(paginatedResponse(map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"name":      map[string]interface{}{"type": "string"},
								"equipment": map[string]interface{}{"type": "string"},
								"pb":        pbSchema,
							},
						},
					})),
				},
			},
			"/lifters/compare": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Compare two lifters side by side",
					"parameters": []map[string]interface{}{
						unitParam,
						{"name": "a", "in": "query", "required": true, "schema": map[string]interface{}{"type": "string"}, "description": "First lifter name"},
						{"name": "b", "in": "query", "required": true, "schema": map[string]interface{}{"type": "string"}, "description": "Second lifter name"},
					},
					"responses": jsonResponse(map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"a": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"lifter": lifterSchema, "stats": careerStatsSchema}},
							"b": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"lifter": lifterSchema, "stats": careerStatsSchema}},
						},
					}),
				},
			},
			"/lifters/{lifterName}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Get a single lifter by name",
					"parameters": []map[string]interface{}{
						{"name": "lifterName", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}},
						unitParam,
					},
					"responses": jsonResponse(lifterSchema),
				},
			},
			"/lifters/{lifterName}/stats": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Career stats for a lifter — meet count, federations, PR progression by equipment",
					"parameters": appendParam([]map[string]interface{}{
						{"name": "lifterName", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}},
					}, unitParam),
					"responses": jsonResponse(careerStatsSchema),
				},
			},
			"/meets": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "List all meets",
					"parameters": concatParams(paginationParams, meetFilterParams, []map[string]interface{}{
						sortParam([]string{"date", "meetName", "federation", "country"}, "date"),
						orderParam,
					}),
					"responses": jsonResponse(paginatedResponse(map[string]interface{}{"type": "array", "items": meetSchema})),
				},
			},
			"/meets/{federationName}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "List meets by federation",
					"parameters": concatParams([]map[string]interface{}{
						{"name": "federationName", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}},
					}, paginationParams, meetFilterParams, []map[string]interface{}{
						sortParam([]string{"date", "meetName", "country"}, "date"),
						orderParam,
					}),
					"responses": jsonResponse(paginatedResponse(map[string]interface{}{"type": "array", "items": meetSchema})),
				},
			},
			"/meets/{federationName}/{meetName}/results": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Get all entries for a specific meet",
					"parameters": concatParams([]map[string]interface{}{
						{"name": "federationName", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}},
						{"name": "meetName", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}},
						unitParam,
					}, paginationParams),
					"responses": jsonResponse(paginatedResponse(map[string]interface{}{"type": "array", "items": meetResultSchema})),
				},
			},
			"/records": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "All-time records per weight class",
					"parameters": concatParams(paginationParams, sexEquipWcParams, []map[string]interface{}{unitParam}),
					"responses": jsonResponse(paginatedResponse(map[string]interface{}{"type": "array", "items": recordSchema})),
				},
			},
			"/federations": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "List all federation names",
					"parameters": concatParams(paginationParams, []map[string]interface{}{
						sortParam([]string{"name"}, "name"),
						orderParam,
					}),
					"responses": jsonResponse(paginatedResponse(map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}})),
				},
			},
			"/openapi.json": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":   "This OpenAPI specification",
					"responses": jsonResponse(map[string]interface{}{"type": "object"}),
				},
			},
		},
	}
}
