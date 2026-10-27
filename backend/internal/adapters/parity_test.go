package adapters_test

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

type expectedTotals struct {
	Invocations       int64                      `json:"invocations"`
	InputTokens       int64                      `json:"input_tokens"`
	OutputTokens      int64                      `json:"output_tokens"`
	CacheReadTokens   int64                      `json:"cache_read_tokens"`
	CacheHitRatePct   float64                    `json:"cache_hit_rate_pct"`
	CostUSD           float64                    `json:"cost_usd"`
	PeakContextTokens int64                      `json:"peak_context_tokens"`
	ContextHealth     domain.ContextHealthStatus `json:"context_health"`
}

type expectedSubagentTurnVelocity struct {
	AvgTurnsPerInvocation float64            `json:"avg_turns_per_invocation"`
	PeakTurns             int64              `json:"peak_turns"`
	ByRole                map[string]float64 `json:"by_role"`
}

type expectedTestToCodeRatio struct {
	Ratio          string `json:"ratio"`
	TestLinesAdded int64  `json:"test_lines_added"`
	ProdLinesAdded int64  `json:"prod_lines_added"`
}

type expectedToolExecution struct {
	TotalCalls           int64   `json:"total_calls"`
	ErrorsAndCorrections int64   `json:"errors_and_corrections"`
	SuccessRatePct       float64 `json:"success_rate_pct"`
}

type expectedContextWindowHealth struct {
	HealthyCeilingTokens int64                      `json:"healthy_ceiling_tokens"`
	Status               domain.ContextHealthStatus `json:"status"`
	RootPeakTokens       int64                      `json:"root_peak_tokens"`
	RootSaturationPct    float64                    `json:"root_saturation_pct"`
	SubagentsPeakTokens  int64                      `json:"subagents_peak_tokens"`
	SubagentsPeakRole    string                     `json:"subagents_peak_role"`
	Compactions          int64                      `json:"compactions"`
	PeakByRole           map[string]int64           `json:"peak_by_role"`
}

type expectedCodeDelivered struct {
	LinesAdded   int64 `json:"lines_added"`
	LinesDeleted int64 `json:"lines_deleted"`
	FilesChanged int64 `json:"files_changed"`
}

type expectedWorkflowEfficiency struct {
	SDLCIterations            int64                        `json:"sdlc_iterations"`
	ActiveDurationSeconds     int64                        `json:"active_duration_seconds"`
	ActiveDurationFormatted   string                       `json:"active_duration_formatted"`
	ReasoningTokens           int64                        `json:"reasoning_tokens"`
	ReasoningRatioPct         float64                      `json:"reasoning_ratio_pct"`
	SubagentTurnVelocity      expectedSubagentTurnVelocity `json:"subagent_turn_velocity"`
	TestToProductionRatio     expectedTestToCodeRatio      `json:"test_to_production_ratio"`
	ToolExecution             expectedToolExecution        `json:"tool_execution"`
	ContextWindowHealth       expectedContextWindowHealth  `json:"context_window_health"`
	CodeDelivered             expectedCodeDelivered        `json:"code_delivered"`
	RoleEffortDistributionPct map[string]float64           `json:"role_effort_distribution_pct"`
}

type expectedTicketMetrics struct {
	Totals             expectedTotals             `json:"totals"`
	WorkflowEfficiency expectedWorkflowEfficiency `json:"workflow_efficiency"`
}

func almostEqual(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

func findParityWorkspacesDir() string {
	if _, currentFile, _, ok := runtime.Caller(0); ok {
		candidate := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "testdata", "parity_workspaces"))
		if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
			return candidate
		}
	}
	candidates := []string{
		filepath.Join("..", "..", "..", "testdata", "parity_workspaces"),
		filepath.Join("..", "..", "testdata", "parity_workspaces"),
		filepath.Join("testdata", "parity_workspaces"),
	}
	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && fi.IsDir() {
			return c
		}
	}
	return filepath.Join("..", "..", "..", "testdata", "parity_workspaces")
}

func TestCrossEngineParityOnWorkspaceLedgers(t *testing.T) {
	workspacesDir := findParityWorkspacesDir()

	report := adapters.ScanWorkspaces(workspacesDir)
	if len(report.Errors) > 0 {
		for _, errStr := range report.Errors {
			t.Logf("Scan warning: %s", errStr)
		}
	}
	if len(report.Ledgers) == 0 {
		t.Fatalf("expected ledgers in %s, got 0", workspacesDir)
	}

	// Python reference outputs extracted from running agentic_lens scan_workspaces on the actual workspace ledgers
	referenceJSON := `{
  "AL-1": {
    "totals": {
      "invocations": 7,
      "input_tokens": 2233835,
      "output_tokens": 76668,
      "cache_read_tokens": 16376447,
      "cache_hit_rate_pct": 88.0,
      "cost_usd": 3.3037,
      "peak_context_tokens": 123350,
      "context_health": "ELEVATED"
    },
    "workflow_efficiency": {
      "sdlc_iterations": 1,
      "active_duration_seconds": 3439,
      "active_duration_formatted": "57m 19s",
      "reasoning_tokens": 30019,
      "reasoning_ratio_pct": 28.1,
      "subagent_turn_velocity": {
        "avg_turns_per_invocation": 23.5,
        "peak_turns": 41,
        "by_role": {
          "coder": 30.5,
          "planner": 17.0,
          "reviewer": 17.5,
          "tester": 28.0
        }
      },
      "test_to_production_ratio": {
        "ratio": "1.0:1",
        "test_lines_added": 803,
        "prod_lines_added": 796
      },
      "tool_execution": {
        "total_calls": 322,
        "errors_and_corrections": 3,
        "success_rate_pct": 99.1
      },
      "context_window_health": {
        "healthy_ceiling_tokens": 200000,
        "status": "ELEVATED",
        "root_peak_tokens": 123350,
        "root_saturation_pct": 61.7,
        "subagents_peak_tokens": 85656,
        "subagents_peak_role": "coder",
        "compactions": 0,
        "peak_by_role": {
          "coder": 85656,
          "general": 123350,
          "planner": 52416,
          "reviewer": 81209,
          "tester": 84421
        }
      },
      "code_delivered": {
        "lines_added": 1599,
        "lines_deleted": 30,
        "files_changed": 16
      },
      "role_effort_distribution_pct": {
        "coder": 43.6,
        "planner": 11.8,
        "reviewer": 23.5,
        "tester": 21.1
      }
    }
  },
  "AL-10": {
    "totals": {
      "invocations": 4,
      "input_tokens": 1063656,
      "output_tokens": 33304,
      "cache_read_tokens": 7414256,
      "cache_hit_rate_pct": 87.5,
      "cost_usd": 1.6576,
      "peak_context_tokens": 87045,
      "context_health": "HEALTHY"
    },
    "workflow_efficiency": {
      "sdlc_iterations": 1,
      "active_duration_seconds": 1999,
      "active_duration_formatted": "33m 19s",
      "reasoning_tokens": 47716,
      "reasoning_ratio_pct": 58.9,
      "subagent_turn_velocity": {
        "avg_turns_per_invocation": 25.7,
        "peak_turns": 31,
        "by_role": {
          "coder": 31.0,
          "planner": 23.0
        }
      },
      "test_to_production_ratio": {
        "ratio": "0:0",
        "test_lines_added": 0,
        "prod_lines_added": 0
      },
      "tool_execution": {
        "total_calls": 177,
        "errors_and_corrections": 8,
        "success_rate_pct": 95.5
      },
      "context_window_health": {
        "healthy_ceiling_tokens": 200000,
        "status": "HEALTHY",
        "root_peak_tokens": 77899,
        "root_saturation_pct": 38.9,
        "subagents_peak_tokens": 87045,
        "subagents_peak_role": "planner",
        "compactions": 0,
        "peak_by_role": {
          "coder": 84570,
          "general": 77899,
          "planner": 87045
        }
      },
      "code_delivered": {
        "lines_added": 0,
        "lines_deleted": 0,
        "files_changed": 0
      },
      "role_effort_distribution_pct": {
        "coder": 40.5,
        "planner": 59.5
      }
    }
  },
  "AL-11": {
    "totals": {
      "invocations": 5,
      "input_tokens": 2349477,
      "output_tokens": 82697,
      "cache_read_tokens": 29777412,
      "cache_hit_rate_pct": 92.7,
      "cost_usd": 4.5215,
      "peak_context_tokens": 179717,
      "context_health": "ELEVATED"
    },
    "workflow_efficiency": {
      "sdlc_iterations": 1,
      "active_duration_seconds": 3460,
      "active_duration_formatted": "57m 40s",
      "reasoning_tokens": 57585,
      "reasoning_ratio_pct": 41.0,
      "subagent_turn_velocity": {
        "avg_turns_per_invocation": 62.8,
        "peak_turns": 106,
        "by_role": {
          "coder": 106.0,
          "planner": 47.0,
          "reviewer": 27.0,
          "tester": 71.0
        }
      },
      "test_to_production_ratio": {
        "ratio": "0.5:1",
        "test_lines_added": 895,
        "prod_lines_added": 1808
      },
      "tool_execution": {
        "total_calls": 402,
        "errors_and_corrections": 6,
        "success_rate_pct": 98.5
      },
      "context_window_health": {
        "healthy_ceiling_tokens": 200000,
        "status": "ELEVATED",
        "root_peak_tokens": 142662,
        "root_saturation_pct": 71.3,
        "subagents_peak_tokens": 179717,
        "subagents_peak_role": "coder",
        "compactions": 0,
        "peak_by_role": {
          "coder": 179717,
          "general": 142662,
          "planner": 94178,
          "reviewer": 75376,
          "tester": 138499
        }
      },
      "code_delivered": {
        "lines_added": 2717,
        "lines_deleted": 13,
        "files_changed": 14
      },
      "role_effort_distribution_pct": {
        "coder": 49.1,
        "planner": 15.7,
        "reviewer": 7.6,
        "tester": 27.6
      }
    }
  },
  "AL-12": {
    "totals": {
      "invocations": 5,
      "input_tokens": 2570848,
      "output_tokens": 86331,
      "cache_read_tokens": 26768589,
      "cache_hit_rate_pct": 91.2,
      "cost_usd": 4.4845,
      "peak_context_tokens": 168532,
      "context_health": "ELEVATED"
    },
    "workflow_efficiency": {
      "sdlc_iterations": 1,
      "active_duration_seconds": 3784,
      "active_duration_formatted": "1h 3m",
      "reasoning_tokens": 59998,
      "reasoning_ratio_pct": 41.0,
      "subagent_turn_velocity": {
        "avg_turns_per_invocation": 42.5,
        "peak_turns": 75,
        "by_role": {
          "coder": 75.0,
          "planner": 23.0,
          "reviewer": 32.0,
          "tester": 40.0
        }
      },
      "test_to_production_ratio": {
        "ratio": "0.9:1",
        "test_lines_added": 870,
        "prod_lines_added": 946
      },
      "tool_execution": {
        "total_calls": 425,
        "errors_and_corrections": 6,
        "success_rate_pct": 98.6
      },
      "context_window_health": {
        "healthy_ceiling_tokens": 200000,
        "status": "ELEVATED",
        "root_peak_tokens": 168532,
        "root_saturation_pct": 84.3,
        "subagents_peak_tokens": 105014,
        "subagents_peak_role": "coder",
        "compactions": 0,
        "peak_by_role": {
          "coder": 105014,
          "general": 168532,
          "planner": 58111,
          "reviewer": 54018,
          "tester": 94236
        }
      },
      "code_delivered": {
        "lines_added": 1816,
        "lines_deleted": 39,
        "files_changed": 20
      },
      "role_effort_distribution_pct": {
        "coder": 47.5,
        "planner": 12.9,
        "reviewer": 14.2,
        "tester": 25.3
      }
    }
  },
  "AL-2": {
    "totals": {
      "invocations": 9,
      "input_tokens": 2506984,
      "output_tokens": 66129,
      "cache_read_tokens": 18196796,
      "cache_hit_rate_pct": 87.9,
      "cost_usd": 3.714,
      "peak_context_tokens": 119146,
      "context_health": "ELEVATED"
    },
    "workflow_efficiency": {
      "sdlc_iterations": 1,
      "active_duration_seconds": 4645,
      "active_duration_formatted": "1h 17m",
      "reasoning_tokens": 58936,
      "reasoning_ratio_pct": 47.1,
      "subagent_turn_velocity": {
        "avg_turns_per_invocation": 33.3,
        "peak_turns": 71,
        "by_role": {
          "coder": 37.0,
          "planner": 22.0,
          "reviewer": 23.5,
          "tester": 43.0
        }
      },
      "test_to_production_ratio": {
        "ratio": "2.2:1",
        "test_lines_added": 1237,
        "prod_lines_added": 574
      },
      "tool_execution": {
        "total_calls": 409,
        "errors_and_corrections": 3,
        "success_rate_pct": 99.3
      },
      "context_window_health": {
        "healthy_ceiling_tokens": 200000,
        "status": "ELEVATED",
        "root_peak_tokens": 91319,
        "root_saturation_pct": 45.7,
        "subagents_peak_tokens": 119146,
        "subagents_peak_role": "tester",
        "compactions": 0,
        "peak_by_role": {
          "coder": 83829,
          "orchestrator": 91319,
          "planner": 53283,
          "reviewer": 87410,
          "tester": 119146
        }
      },
      "code_delivered": {
        "lines_added": 1811,
        "lines_deleted": 0,
        "files_changed": 16
      },
      "role_effort_distribution_pct": {
        "coder": 36.0,
        "planner": 7.6,
        "reviewer": 20.4,
        "tester": 36.0
      }
    }
  },
  "AL-3": {
    "totals": {
      "invocations": 9,
      "input_tokens": 2999956,
      "output_tokens": 93869,
      "cache_read_tokens": 22428605,
      "cache_hit_rate_pct": 88.2,
      "cost_usd": 4.4221,
      "peak_context_tokens": 147223,
      "context_health": "ELEVATED"
    },
    "workflow_efficiency": {
      "sdlc_iterations": 1,
      "active_duration_seconds": 4236,
      "active_duration_formatted": "1h 10m",
      "reasoning_tokens": 36787,
      "reasoning_ratio_pct": 28.2,
      "subagent_turn_velocity": {
        "avg_turns_per_invocation": 33.0,
        "peak_turns": 65,
        "by_role": {
          "coder": 36.3,
          "planner": 18.0,
          "reviewer": 52.0,
          "tester": 33.5
        }
      },
      "test_to_production_ratio": {
        "ratio": "0.6:1",
        "test_lines_added": 1112,
        "prod_lines_added": 1868
      },
      "tool_execution": {
        "total_calls": 425,
        "errors_and_corrections": 6,
        "success_rate_pct": 98.6
      },
      "context_window_health": {
        "healthy_ceiling_tokens": 200000,
        "status": "ELEVATED",
        "root_peak_tokens": 147223,
        "root_saturation_pct": 73.6,
        "subagents_peak_tokens": 84526,
        "subagents_peak_role": "reviewer",
        "compactions": 0,
        "peak_by_role": {
          "coder": 79596,
          "orchestrator": 147223,
          "planner": 55575,
          "reviewer": 84526,
          "tester": 73382
        }
      },
      "code_delivered": {
        "lines_added": 2980,
        "lines_deleted": 7,
        "files_changed": 26
      },
      "role_effort_distribution_pct": {
        "coder": 39.8,
        "planner": 13.5,
        "reviewer": 20.2,
        "tester": 26.4
      }
    }
  },
  "AL-4": {
    "totals": {
      "invocations": 10,
      "input_tokens": 6436939,
      "output_tokens": 223985,
      "cache_read_tokens": 55637832,
      "cache_hit_rate_pct": 89.6,
      "cost_usd": 10.5715,
      "peak_context_tokens": 181811,
      "context_health": "ELEVATED"
    },
    "workflow_efficiency": {
      "sdlc_iterations": 2,
      "active_duration_seconds": 12406,
      "active_duration_formatted": "3h 26m",
      "reasoning_tokens": 194925,
      "reasoning_ratio_pct": 46.5,
      "subagent_turn_velocity": {
        "avg_turns_per_invocation": 68.4,
        "peak_turns": 188,
        "by_role": {
          "coder": 93.0,
          "planner": 33.0,
          "reviewer": 44.5,
          "tester": 91.0
        }
      },
      "test_to_production_ratio": {
        "ratio": "0.6:1",
        "test_lines_added": 3170,
        "prod_lines_added": 5197
      },
      "tool_execution": {
        "total_calls": 878,
        "errors_and_corrections": 16,
        "success_rate_pct": 98.2
      },
      "context_window_health": {
        "healthy_ceiling_tokens": 200000,
        "status": "ELEVATED",
        "root_peak_tokens": 172386,
        "root_saturation_pct": 86.2,
        "subagents_peak_tokens": 181811,
        "subagents_peak_role": "coder",
        "compactions": 0,
        "peak_by_role": {
          "coder": 181811,
          "orchestrator": 172386,
          "planner": 94064,
          "reviewer": 128153,
          "tester": 166685
        }
      },
      "code_delivered": {
        "lines_added": 8367,
        "lines_deleted": 86,
        "files_changed": 34
      },
      "role_effort_distribution_pct": {
        "coder": 48.0,
        "planner": 8.8,
        "reviewer": 13.4,
        "tester": 29.8
      }
    }
  },
  "GEN-1": {
    "totals": {
      "invocations": 1,
      "input_tokens": 172617,
      "output_tokens": 3999,
      "cache_read_tokens": 333076,
      "cache_hit_rate_pct": 65.9,
      "cost_usd": 0.1908,
      "peak_context_tokens": 26252,
      "context_health": "HEALTHY"
    },
    "workflow_efficiency": {
      "sdlc_iterations": 1,
      "active_duration_seconds": 169,
      "active_duration_formatted": "2m 49s",
      "reasoning_tokens": 5700,
      "reasoning_ratio_pct": 58.8,
      "subagent_turn_velocity": {
        "avg_turns_per_invocation": 0.0,
        "peak_turns": 0,
        "by_role": {}
      },
      "test_to_production_ratio": {
        "ratio": "0:0",
        "test_lines_added": 0,
        "prod_lines_added": 0
      },
      "tool_execution": {
        "total_calls": 25,
        "errors_and_corrections": 0,
        "success_rate_pct": 100.0
      },
      "context_window_health": {
        "healthy_ceiling_tokens": 200000,
        "status": "HEALTHY",
        "root_peak_tokens": 26252,
        "root_saturation_pct": 13.1,
        "subagents_peak_tokens": 0,
        "subagents_peak_role": "none",
        "compactions": 0,
        "peak_by_role": {
          "general": 26252
        }
      },
      "code_delivered": {
        "lines_added": 0,
        "lines_deleted": 0,
        "files_changed": 0
      },
      "role_effort_distribution_pct": {}
    }
  },
  "GEN-2": {
    "totals": {
      "invocations": 1,
      "input_tokens": 131392,
      "output_tokens": 3036,
      "cache_read_tokens": 512071,
      "cache_hit_rate_pct": 79.6,
      "cost_usd": 0.1662,
      "peak_context_tokens": 29125,
      "context_health": "HEALTHY"
    },
    "workflow_efficiency": {
      "sdlc_iterations": 1,
      "active_duration_seconds": 108,
      "active_duration_formatted": "1m 48s",
      "reasoning_tokens": 4772,
      "reasoning_ratio_pct": 61.1,
      "subagent_turn_velocity": {
        "avg_turns_per_invocation": 0.0,
        "peak_turns": 0,
        "by_role": {}
      },
      "test_to_production_ratio": {
        "ratio": "0:0",
        "test_lines_added": 0,
        "prod_lines_added": 0
      },
      "tool_execution": {
        "total_calls": 29,
        "errors_and_corrections": 0,
        "success_rate_pct": 100.0
      },
      "context_window_health": {
        "healthy_ceiling_tokens": 200000,
        "status": "HEALTHY",
        "root_peak_tokens": 29125,
        "root_saturation_pct": 14.6,
        "subagents_peak_tokens": 0,
        "subagents_peak_role": "none",
        "compactions": 0,
        "peak_by_role": {
          "general": 29125
        }
      },
      "code_delivered": {
        "lines_added": 0,
        "lines_deleted": 0,
        "files_changed": 0
      },
      "role_effort_distribution_pct": {}
    }
  },
  "PO-1": {
    "totals": {
      "invocations": 1,
      "input_tokens": 1139501,
      "output_tokens": 28249,
      "cache_read_tokens": 7633800,
      "cache_hit_rate_pct": 87.0,
      "cost_usd": 1.6085,
      "peak_context_tokens": 113824,
      "context_health": "ELEVATED"
    },
    "workflow_efficiency": {
      "sdlc_iterations": 1,
      "active_duration_seconds": 1234,
      "active_duration_formatted": "20m 34s",
      "reasoning_tokens": 20118,
      "reasoning_ratio_pct": 41.6,
      "subagent_turn_velocity": {
        "avg_turns_per_invocation": 0.0,
        "peak_turns": 0,
        "by_role": {}
      },
      "test_to_production_ratio": {
        "ratio": "0:0",
        "test_lines_added": 0,
        "prod_lines_added": 0
      },
      "tool_execution": {
        "total_calls": 121,
        "errors_and_corrections": 1,
        "success_rate_pct": 99.2
      },
      "context_window_health": {
        "healthy_ceiling_tokens": 200000,
        "status": "ELEVATED",
        "root_peak_tokens": 113824,
        "root_saturation_pct": 56.9,
        "subagents_peak_tokens": 0,
        "subagents_peak_role": "none",
        "compactions": 0,
        "peak_by_role": {
          "general": 113824
        }
      },
      "code_delivered": {
        "lines_added": 0,
        "lines_deleted": 0,
        "files_changed": 0
      },
      "role_effort_distribution_pct": {}
    }
  },
  "PO-2": {
    "totals": {
      "invocations": 1,
      "input_tokens": 757331,
      "output_tokens": 17281,
      "cache_read_tokens": 6092711,
      "cache_hit_rate_pct": 88.9,
      "cost_usd": 1.1545,
      "peak_context_tokens": 88541,
      "context_health": "HEALTHY"
    },
    "workflow_efficiency": {
      "sdlc_iterations": 1,
      "active_duration_seconds": 1241,
      "active_duration_formatted": "20m 41s",
      "reasoning_tokens": 17265,
      "reasoning_ratio_pct": 50.0,
      "subagent_turn_velocity": {
        "avg_turns_per_invocation": 0.0,
        "peak_turns": 0,
        "by_role": {}
      },
      "test_to_production_ratio": {
        "ratio": "0:0",
        "test_lines_added": 0,
        "prod_lines_added": 0
      },
      "tool_execution": {
        "total_calls": 110,
        "errors_and_corrections": 1,
        "success_rate_pct": 99.1
      },
      "context_window_health": {
        "healthy_ceiling_tokens": 200000,
        "status": "HEALTHY",
        "root_peak_tokens": 88541,
        "root_saturation_pct": 44.3,
        "subagents_peak_tokens": 0,
        "subagents_peak_role": "none",
        "compactions": 0,
        "peak_by_role": {
          "general": 88541
        }
      },
      "code_delivered": {
        "lines_added": 0,
        "lines_deleted": 0,
        "files_changed": 0
      },
      "role_effort_distribution_pct": {}
    }
  },
  "PO-3": {
    "totals": {
      "invocations": 1,
      "input_tokens": 490913,
      "output_tokens": 14833,
      "cache_read_tokens": 3248322,
      "cache_hit_rate_pct": 86.9,
      "cost_usd": 0.7481,
      "peak_context_tokens": 86075,
      "context_health": "HEALTHY"
    },
    "workflow_efficiency": {
      "sdlc_iterations": 1,
      "active_duration_seconds": 592,
      "active_duration_formatted": "9m 52s",
      "reasoning_tokens": 21515,
      "reasoning_ratio_pct": 59.2,
      "subagent_turn_velocity": {
        "avg_turns_per_invocation": 0.0,
        "peak_turns": 0,
        "by_role": {}
      },
      "test_to_production_ratio": {
        "ratio": "0:0",
        "test_lines_added": 0,
        "prod_lines_added": 0
      },
      "tool_execution": {
        "total_calls": 71,
        "errors_and_corrections": 0,
        "success_rate_pct": 100.0
      },
      "context_window_health": {
        "healthy_ceiling_tokens": 200000,
        "status": "HEALTHY",
        "root_peak_tokens": 86075,
        "root_saturation_pct": 43.0,
        "subagents_peak_tokens": 0,
        "subagents_peak_role": "none",
        "compactions": 0,
        "peak_by_role": {
          "general": 86075
        }
      },
      "code_delivered": {
        "lines_added": 0,
        "lines_deleted": 0,
        "files_changed": 0
      },
      "role_effort_distribution_pct": {}
    }
  }
}`

	var expected map[string]expectedTicketMetrics
	if err := json.Unmarshal([]byte(referenceJSON), &expected); err != nil {
		t.Fatalf("failed unmarshaling reference JSON: %v", err)
	}

	for _, ledger := range report.Ledgers {
		if ledger.TicketID == nil {
			continue
		}
		ticketID := *ledger.TicketID
		exp, ok := expected[ticketID]
		if !ok {
			// Skip any workspaces outside the reference set
			continue
		}

		t.Run(ticketID, func(t *testing.T) {
			tot := ledger.Totals
			eff := ledger.WorkflowEfficiency

			// 1. Totals parity
			if tot.Invocations != exp.Totals.Invocations {
				t.Errorf("[%s] Invocations: expected %d, got %d", ticketID, exp.Totals.Invocations, tot.Invocations)
			}
			if tot.InputTokens != exp.Totals.InputTokens {
				t.Errorf("[%s] InputTokens: expected %d, got %d", ticketID, exp.Totals.InputTokens, tot.InputTokens)
			}
			if tot.OutputTokens != exp.Totals.OutputTokens {
				t.Errorf("[%s] OutputTokens: expected %d, got %d", ticketID, exp.Totals.OutputTokens, tot.OutputTokens)
			}
			if tot.CacheReadTokens != exp.Totals.CacheReadTokens {
				t.Errorf("[%s] CacheReadTokens: expected %d, got %d", ticketID, exp.Totals.CacheReadTokens, tot.CacheReadTokens)
			}
			if !almostEqual(tot.CacheHitRatePct, exp.Totals.CacheHitRatePct, 0.05) {
				t.Errorf("[%s] CacheHitRatePct: expected %f, got %f", ticketID, exp.Totals.CacheHitRatePct, tot.CacheHitRatePct)
			}
			if !almostEqual(tot.CostUSD, exp.Totals.CostUSD, 0.0001) {
				t.Errorf("[%s] CostUSD: expected %f, got %f", ticketID, exp.Totals.CostUSD, tot.CostUSD)
			}
			if tot.PeakContextTokens != exp.Totals.PeakContextTokens {
				t.Errorf("[%s] PeakContextTokens: expected %d, got %d", ticketID, exp.Totals.PeakContextTokens, tot.PeakContextTokens)
			}
			if tot.ContextHealth != exp.Totals.ContextHealth {
				t.Errorf("[%s] ContextHealth: expected %s, got %s", ticketID, exp.Totals.ContextHealth, tot.ContextHealth)
			}

			// 2. Workflow Efficiency parity
			if eff.SDLCIterations != exp.WorkflowEfficiency.SDLCIterations {
				t.Errorf("[%s] SDLCIterations: expected %d, got %d", ticketID, exp.WorkflowEfficiency.SDLCIterations, eff.SDLCIterations)
			}
			if eff.ActiveDurationSeconds != exp.WorkflowEfficiency.ActiveDurationSeconds {
				t.Errorf("[%s] ActiveDurationSeconds: expected %d, got %d", ticketID, exp.WorkflowEfficiency.ActiveDurationSeconds, eff.ActiveDurationSeconds)
			}
			if eff.ActiveDurationFormatted != exp.WorkflowEfficiency.ActiveDurationFormatted {
				t.Errorf("[%s] ActiveDurationFormatted: expected %s, got %s", ticketID, exp.WorkflowEfficiency.ActiveDurationFormatted, eff.ActiveDurationFormatted)
			}
			if eff.ReasoningTokens != exp.WorkflowEfficiency.ReasoningTokens {
				t.Errorf("[%s] ReasoningTokens: expected %d, got %d", ticketID, exp.WorkflowEfficiency.ReasoningTokens, eff.ReasoningTokens)
			}
			if !almostEqual(eff.ReasoningRatioPct, exp.WorkflowEfficiency.ReasoningRatioPct, 0.05) {
				t.Errorf("[%s] ReasoningRatioPct: expected %f, got %f", ticketID, exp.WorkflowEfficiency.ReasoningRatioPct, eff.ReasoningRatioPct)
			}

			// Subagent turn velocity
			satv := eff.SubagentTurnVelocity
			expSatv := exp.WorkflowEfficiency.SubagentTurnVelocity
			if !almostEqual(satv.AvgTurnsPerInvocation, expSatv.AvgTurnsPerInvocation, 0.05) {
				t.Errorf("[%s] AvgTurnsPerInvocation: expected %f, got %f", ticketID, expSatv.AvgTurnsPerInvocation, satv.AvgTurnsPerInvocation)
			}
			if satv.PeakTurns != expSatv.PeakTurns {
				t.Errorf("[%s] PeakTurns: expected %d, got %d", ticketID, expSatv.PeakTurns, satv.PeakTurns)
			}
			for role, expVal := range expSatv.ByRole {
				gotVal, ok := satv.ByRole[domain.AgentRole(role)]
				if !ok || !almostEqual(gotVal, expVal, 0.05) {
					t.Errorf("[%s] ByRole[%s]: expected %f, got %f", ticketID, role, expVal, gotVal)
				}
			}

			// Test to production ratio
			ttpr := eff.TestToProductionRatio
			expTtpr := exp.WorkflowEfficiency.TestToProductionRatio
			if ttpr.Ratio != expTtpr.Ratio {
				t.Errorf("[%s] TestToProd Ratio: expected %s, got %s", ticketID, expTtpr.Ratio, ttpr.Ratio)
			}
			if ttpr.TestLinesAdded != expTtpr.TestLinesAdded {
				t.Errorf("[%s] TestLinesAdded: expected %d, got %d", ticketID, expTtpr.TestLinesAdded, ttpr.TestLinesAdded)
			}
			if ttpr.ProdLinesAdded != expTtpr.ProdLinesAdded {
				t.Errorf("[%s] ProdLinesAdded: expected %d, got %d", ticketID, expTtpr.ProdLinesAdded, ttpr.ProdLinesAdded)
			}

			// Tool execution
			te := eff.ToolExecution
			expTe := exp.WorkflowEfficiency.ToolExecution
			if te.TotalCalls != expTe.TotalCalls {
				t.Errorf("[%s] TotalCalls: expected %d, got %d", ticketID, expTe.TotalCalls, te.TotalCalls)
			}
			if te.ErrorsAndCorrections != expTe.ErrorsAndCorrections {
				t.Errorf("[%s] ErrorsAndCorrections: expected %d, got %d", ticketID, expTe.ErrorsAndCorrections, te.ErrorsAndCorrections)
			}
			if !almostEqual(te.SuccessRatePct, expTe.SuccessRatePct, 0.05) {
				t.Errorf("[%s] SuccessRatePct: expected %f, got %f", ticketID, expTe.SuccessRatePct, te.SuccessRatePct)
			}

			// Context window health
			cwh := eff.ContextWindowHealth
			expCwh := exp.WorkflowEfficiency.ContextWindowHealth
			if cwh.Status != expCwh.Status {
				t.Errorf("[%s] CWH Status: expected %s, got %s", ticketID, expCwh.Status, cwh.Status)
			}
			if cwh.RootPeakTokens != expCwh.RootPeakTokens {
				t.Errorf("[%s] RootPeakTokens: expected %d, got %d", ticketID, expCwh.RootPeakTokens, cwh.RootPeakTokens)
			}
			if !almostEqual(cwh.RootSaturationPct, expCwh.RootSaturationPct, 0.05) {
				t.Errorf("[%s] RootSaturationPct: expected %f, got %f", ticketID, expCwh.RootSaturationPct, cwh.RootSaturationPct)
			}
			if cwh.SubagentsPeakTokens != expCwh.SubagentsPeakTokens {
				t.Errorf("[%s] SubagentsPeakTokens: expected %d, got %d", ticketID, expCwh.SubagentsPeakTokens, cwh.SubagentsPeakTokens)
			}
			if string(cwh.SubagentsPeakRole) != expCwh.SubagentsPeakRole {
				t.Errorf("[%s] SubagentsPeakRole: expected %s, got %s", ticketID, expCwh.SubagentsPeakRole, cwh.SubagentsPeakRole)
			}
			for role, expPeak := range expCwh.PeakByRole {
				gotPeak, ok := cwh.PeakByRole[domain.AgentRole(role)]
				if !ok || gotPeak != expPeak {
					t.Errorf("[%s] PeakByRole[%s]: expected %d, got %d", ticketID, role, expPeak, gotPeak)
				}
			}

			// Code delivered
			cd := eff.CodeDelivered
			expCd := exp.WorkflowEfficiency.CodeDelivered
			if cd.LinesAdded != expCd.LinesAdded {
				t.Errorf("[%s] LinesAdded: expected %d, got %d", ticketID, expCd.LinesAdded, cd.LinesAdded)
			}
			if cd.LinesDeleted != expCd.LinesDeleted {
				t.Errorf("[%s] LinesDeleted: expected %d, got %d", ticketID, expCd.LinesDeleted, cd.LinesDeleted)
			}
			if cd.FilesChanged != expCd.FilesChanged {
				t.Errorf("[%s] FilesChanged: expected %d, got %d", ticketID, expCd.FilesChanged, cd.FilesChanged)
			}

			// Role effort distribution
			for role, expDist := range exp.WorkflowEfficiency.RoleEffortDistributionPct {
				gotDist, ok := eff.RoleEffortDistributionPct[role]
				if !ok || !almostEqual(gotDist, expDist, 0.05) {
					t.Errorf("[%s] RoleEffort[%s]: expected %f, got %f", ticketID, role, expDist, gotDist)
				}
			}
		})
	}
}

func TestFleetSummaryAndComparisonParity(t *testing.T) {
	workspacesDir := findParityWorkspacesDir()

	report := adapters.ScanWorkspaces(workspacesDir)
	referenceTickets := map[string]bool{
		"AL-1": true, "AL-10": true, "AL-11": true, "AL-12": true,
		"AL-2": true, "AL-3": true, "AL-4": true, "GEN-1": true,
		"GEN-2": true, "PO-1": true, "PO-2": true, "PO-3": true,
	}
	var refLedgers []domain.AgenticMetricsLedger
	for _, l := range report.Ledgers {
		if l.TicketID != nil && referenceTickets[*l.TicketID] {
			refLedgers = append(refLedgers, l)
		}
	}
	fleet := domain.CalculateFleetSummary(refLedgers)

	// Reference values from Python reference run
	if fleet.TotalTickets != 12 {
		t.Errorf("TotalTickets: expected 12, got %d", fleet.TotalTickets)
	}
	if !almostEqual(fleet.TotalCostUSD, 36.5430, 0.0001) {
		t.Errorf("TotalCostUSD: expected 36.5430, got %f", fleet.TotalCostUSD)
	}
	if fleet.TotalInvocations != 54 {
		t.Errorf("TotalInvocations: expected 54, got %d", fleet.TotalInvocations)
	}
	if fleet.TotalInputTokens != 22853449 {
		t.Errorf("TotalInputTokens: expected 22853449, got %d", fleet.TotalInputTokens)
	}
	if fleet.TotalOutputTokens != 730381 {
		t.Errorf("TotalOutputTokens: expected 730381, got %d", fleet.TotalOutputTokens)
	}
	if fleet.TotalCacheReadTokens != 194419917 {
		t.Errorf("TotalCacheReadTokens: expected 194419917, got %d", fleet.TotalCacheReadTokens)
	}
	if !almostEqual(fleet.OverallCacheHitRatePct, 89.5, 0.05) {
		t.Errorf("OverallCacheHitRatePct: expected 89.5, got %f", fleet.OverallCacheHitRatePct)
	}
	if fleet.TotalDurationSeconds != 37313 {
		t.Errorf("TotalDurationSeconds: expected 37313, got %d", fleet.TotalDurationSeconds)
	}
	if fleet.FormattedTotalDuration != "10h 21m" {
		t.Errorf("FormattedTotalDuration: expected '10h 21m', got %s", fleet.FormattedTotalDuration)
	}
	if fleet.TotalLinesAdded != 19290 {
		t.Errorf("TotalLinesAdded: expected 19290, got %d", fleet.TotalLinesAdded)
	}
	if fleet.TotalLinesDeleted != 175 {
		t.Errorf("TotalLinesDeleted: expected 175, got %d", fleet.TotalLinesDeleted)
	}
	if fleet.TotalFilesChanged != 126 {
		t.Errorf("TotalFilesChanged: expected 126, got %d", fleet.TotalFilesChanged)
	}
	if fleet.TotalReasoningTokens != 555336 {
		t.Errorf("TotalReasoningTokens: expected 555336, got %d", fleet.TotalReasoningTokens)
	}
	if !almostEqual(fleet.OverallReasoningRatioPct, 43.2, 0.05) {
		t.Errorf("OverallReasoningRatioPct: expected 43.2, got %f", fleet.OverallReasoningRatioPct)
	}

	expectedRoleCosts := map[string]float64{
		"orchestrator": 7.8645,
		"general":      10.9622,
		"planner":      2.3555,
		"coder":        7.9505,
		"tester":       4.8435,
		"reviewer":     2.5671,
	}
	for role, expCost := range expectedRoleCosts {
		gotCost, ok := fleet.RoleCosts[domain.AgentRole(role)]
		if !ok || !almostEqual(gotCost, expCost, 0.0001) {
			t.Errorf("RoleCost[%s]: expected %f, got %f", role, expCost, gotCost)
		}
	}

	modelCost, ok := fleet.ModelCosts["gemini-3.8-flash"]
	if !ok || !almostEqual(modelCost, 36.5430, 0.0001) {
		t.Errorf("ModelCost[gemini-3.8-flash]: expected 36.5430, got %f", modelCost)
	}
}

func TestStructuralClassificationParityOnWorkspaces(t *testing.T) {
	workspacesDir := findParityWorkspacesDir()
	report := adapters.ScanWorkspaces(workspacesDir)
	if len(report.Ledgers) == 0 {
		t.Fatalf("expected ledgers in %s, got 0", workspacesDir)
	}

	for _, ledger := range report.Ledgers {
		if ledger.TicketID == nil {
			continue
		}
		ticketID := *ledger.TicketID
		t.Run(ticketID, func(t *testing.T) {
			facts := ledger.Facts
			var hasRoot bool
			var expRootPeak int64
			var hasChild bool
			var expChildPeak int64

			for _, inv := range facts.InvocationFacts() {
				if inv.ParentID == nil {
					hasRoot = true
					if inv.Tokens.PeakContext > expRootPeak {
						expRootPeak = inv.Tokens.PeakContext
					}
				} else {
					hasChild = true
					if inv.Tokens.PeakContext > expChildPeak {
						expChildPeak = inv.Tokens.PeakContext
					}
				}
			}

			cwh := ledger.WorkflowEfficiency.ContextWindowHealth
			if hasRoot {
				if cwh.RootPeakTokens != expRootPeak {
					t.Errorf("[%s] RootPeakTokens: expected %d, got %d", ticketID, expRootPeak, cwh.RootPeakTokens)
				}
			} else {
				if cwh.RootPeakTokens != 0 {
					t.Errorf("[%s] RootPeakTokens: expected 0, got %d", ticketID, cwh.RootPeakTokens)
				}
			}

			if !hasChild {
				if cwh.SubagentsPeakTokens != 0 {
					t.Errorf("[%s] expected 0 SubagentsPeakTokens, got %d", ticketID, cwh.SubagentsPeakTokens)
				}
				if cwh.SubagentsPeakRole != "none" {
					t.Errorf("[%s] expected SubagentsPeakRole 'none', got %s", ticketID, cwh.SubagentsPeakRole)
				}
				if len(ledger.WorkflowEfficiency.RoleEffortDistributionPct) != 0 {
					t.Errorf("[%s] expected empty RoleEffortDistributionPct, got %+v", ticketID, ledger.WorkflowEfficiency.RoleEffortDistributionPct)
				}
			} else {
				if cwh.SubagentsPeakTokens != expChildPeak {
					t.Errorf("[%s] SubagentsPeakTokens: expected %d, got %d", ticketID, expChildPeak, cwh.SubagentsPeakTokens)
				}
			}
		})
	}
}
