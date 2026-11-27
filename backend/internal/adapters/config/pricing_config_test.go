package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func TestModelPricingYAMLValidationAndRoundTrip(t *testing.T) {
	baseBody := fixtureSharedBody(t)

	t.Run("absent model_pricing is valid and disables estimation", func(t *testing.T) {
		cfg, err := loadBody(t, baseBody)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cfg.ModelPricing.Rates) != 0 {
			t.Fatalf("expected empty rates, got %+v", cfg.ModelPricing.Rates)
		}
		if cfg.ModelPricing.EffectiveCatalogID() != "" {
			t.Fatalf("expected empty catalog ID, got %q", cfg.ModelPricing.EffectiveCatalogID())
		}
	})

	t.Run("explicit empty rates list is valid", func(t *testing.T) {
		body := baseBody + "\nmodel_pricing:\n  rates: []\n"
		cfg, err := loadBody(t, body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cfg.ModelPricing.Rates) != 0 {
			t.Fatalf("expected 0 rates, got %+v", cfg.ModelPricing.Rates)
		}
	})

	t.Run("valid model_pricing.rates loads exact and glob rules", func(t *testing.T) {
		body := baseBody + `
model_pricing:
  rates:
    - model: gpt-5.4
      input_per_million: 2.0
      cache_read_per_million: 0.25
      output_per_million: 8.0
    - model: "*gemini-3.8*"
      input_per_million: 0
      cache_read_per_million: 0
      output_per_million: 0
`
		cfg, err := loadBody(t, body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cfg.ModelPricing.Rates) != 2 {
			t.Fatalf("got %d rates, want 2", len(cfg.ModelPricing.Rates))
		}
		if cfg.ModelPricing.Rates[0].Model != "gpt-5.4" ||
			cfg.ModelPricing.Rates[0].InputPerMillion != 2.0 ||
			cfg.ModelPricing.Rates[0].CacheReadPerMillion != 0.25 ||
			cfg.ModelPricing.Rates[0].OutputPerMillion != 8.0 {
			t.Fatalf("unexpected first rate: %+v", cfg.ModelPricing.Rates[0])
		}
		if cfg.ModelPricing.Rates[1].Model != "*gemini-3.8*" || cfg.ModelPricing.Rates[1].OutputPerMillion != 0 {
			t.Fatalf("unexpected second rate: %+v", cfg.ModelPricing.Rates[1])
		}
		if !strings.HasPrefix(cfg.ModelPricing.EffectiveCatalogID(), "sha256:") {
			t.Fatalf("expected sha256 catalog ID, got %q", cfg.ModelPricing.EffectiveCatalogID())
		}
	})

	invalidCases := []struct {
		name    string
		snippet string
		wantErr string
	}{
		{
			name:    "non-mapping model_pricing",
			snippet: "\nmodel_pricing: []\n",
			wantErr: "model_pricing",
		},
		{
			name:    "unsupported field in model_pricing",
			snippet: "\nmodel_pricing:\n  currency: USD\n",
			wantErr: "unsupported field",
		},
		{
			name:    "non-sequence rates",
			snippet: "\nmodel_pricing:\n  rates: {}\n",
			wantErr: "model_pricing.rates",
		},
		{
			name: "string rate rejected",
			snippet: `
model_pricing:
  rates:
    - model: gpt-5.4
      input_per_million: "2.0"
      cache_read_per_million: 0.25
      output_per_million: 8.0
`,
			wantErr: "input_per_million",
		},
		{
			name: "bool rate rejected",
			snippet: `
model_pricing:
  rates:
    - model: gpt-5.4
      input_per_million: 2.0
      cache_read_per_million: true
      output_per_million: 8.0
`,
			wantErr: "cache_read_per_million",
		},
		{
			name: "null rate rejected",
			snippet: `
model_pricing:
  rates:
    - model: gpt-5.4
      input_per_million: 2.0
      cache_read_per_million: 0.25
      output_per_million: null
`,
			wantErr: "output_per_million",
		},
		{
			name: "negative rate rejected",
			snippet: `
model_pricing:
  rates:
    - model: gpt-5.4
      input_per_million: -1.0
      cache_read_per_million: 0.25
      output_per_million: 8.0
`,
			wantErr: "input_per_million",
		},
		{
			name: "NaN YAML rate rejected",
			snippet: `
model_pricing:
  rates:
    - model: gpt-5.4
      input_per_million: .nan
      cache_read_per_million: 0.25
      output_per_million: 8.0
`,
			wantErr: "input_per_million",
		},
		{
			name: "Inf YAML rate rejected",
			snippet: `
model_pricing:
  rates:
    - model: gpt-5.4
      input_per_million: 2.0
      cache_read_per_million: 0.25
      output_per_million: .inf
`,
			wantErr: "output_per_million",
		},
		{
			name: "duplicate normalized rules rejected",
			snippet: `
model_pricing:
  rates:
    - model: GPT-5.4
      input_per_million: 2.0
      cache_read_per_million: 0.25
      output_per_million: 8.0
    - model: gpt-5.4
      input_per_million: 3.0
      cache_read_per_million: 0.5
      output_per_million: 9.0
`,
			wantErr: "duplicates normalized rule",
		},
		{
			name: "invalid glob rejected",
			snippet: `
model_pricing:
  rates:
    - model: "gpt-[5]"
      input_per_million: 2.0
      cache_read_per_million: 0.25
      output_per_million: 8.0
`,
			wantErr: "invalid glob syntax",
		},
		{
			name: "missing model rejected",
			snippet: `
model_pricing:
  rates:
    - input_per_million: 2.0
      cache_read_per_million: 0.25
      output_per_million: 8.0
`,
			wantErr: "model is required",
		},
		{
			name: "missing input_per_million rejected",
			snippet: `
model_pricing:
  rates:
    - model: gpt-5.4
      cache_read_per_million: 0.25
      output_per_million: 8.0
`,
			wantErr: "input_per_million is required",
		},
		{
			name: "missing cache_read_per_million rejected",
			snippet: `
model_pricing:
  rates:
    - model: gpt-5.4
      input_per_million: 2.0
      output_per_million: 8.0
`,
			wantErr: "cache_read_per_million is required",
		},
		{
			name: "missing output_per_million rejected",
			snippet: `
model_pricing:
  rates:
    - model: gpt-5.4
      input_per_million: 2.0
      cache_read_per_million: 0.25
`,
			wantErr: "output_per_million is required",
		},
		{
			name: "duplicate model key in rate item rejected",
			snippet: `
model_pricing:
  rates:
    - model: gpt-5.4
      model: gpt-5.4-mini
      input_per_million: 2.0
      cache_read_per_million: 0.25
      output_per_million: 8.0
`,
			wantErr: "duplicate capture configuration key model",
		},
		{
			name: "duplicate input_per_million key in rate item rejected",
			snippet: `
model_pricing:
  rates:
    - model: gpt-5.4
      input_per_million: 2.0
      input_per_million: 3.0
      cache_read_per_million: 0.25
      output_per_million: 8.0
`,
			wantErr: "duplicate capture configuration key input_per_million",
		},
		{
			name: "duplicate cache_read_per_million key in rate item rejected",
			snippet: `
model_pricing:
  rates:
    - model: gpt-5.4
      input_per_million: 2.0
      cache_read_per_million: 0.25
      cache_read_per_million: 0.50
      output_per_million: 8.0
`,
			wantErr: "duplicate capture configuration key cache_read_per_million",
		},
		{
			name: "duplicate output_per_million key in rate item rejected",
			snippet: `
model_pricing:
  rates:
    - model: gpt-5.4
      input_per_million: 2.0
      cache_read_per_million: 0.25
      output_per_million: 8.0
      output_per_million: 9.0
`,
			wantErr: "duplicate capture configuration key output_per_million",
		},
		{
			name: "duplicate rates key in model_pricing rejected",
			snippet: `
model_pricing:
  rates: []
  rates: []
`,
			wantErr: "duplicate capture configuration key rates",
		},
		{
			name: "unsupported field in rate item rejected",
			snippet: `
model_pricing:
  rates:
    - model: gpt-5.4
      input_per_million: 2.0
      cache_read_per_million: 0.25
      output_per_million: 8.0
      cache_write_per_million: 3.0
`,
			wantErr: "unsupported field",
		},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := loadBody(t, baseBody+tc.snippet)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("loadBody(%s) error = %v, want substring %q", tc.name, err, tc.wantErr)
			}
		})
	}
}

func TestGenerateConfigYAMLPreservesAndOmitsModelPricing(t *testing.T) {
	t.Run("omits model_pricing when empty", func(t *testing.T) {
		cfg := testProjectConfig()
		cfg.ModelPricing = domain.ModelPricingConfig{}
		data, err := config.GenerateConfigYAML(cfg)
		if err != nil {
			t.Fatalf("GenerateConfigYAML error: %v", err)
		}
		if strings.Contains(string(data), "model_pricing") {
			t.Fatalf("expected model_pricing to be omitted when empty, got:\n%s", string(data))
		}
	})

	t.Run("preserves model_pricing across write and reload", func(t *testing.T) {
		cfg := testProjectConfig()
		cfg.ModelPricing = domain.ModelPricingConfig{
			Rates: []domain.ModelPricingRate{
				{Model: "gpt-5.4", InputPerMillion: 2.0, CacheReadPerMillion: 0.25, OutputPerMillion: 8.0},
				{Model: "*gemini-3.8*", InputPerMillion: 0.5, CacheReadPerMillion: 0.05, OutputPerMillion: 1.5},
			},
		}
		data, err := config.GenerateConfigYAML(cfg)
		if err != nil {
			t.Fatalf("GenerateConfigYAML error: %v", err)
		}
		if !strings.Contains(string(data), "model_pricing:") {
			t.Fatalf("expected model_pricing block in generated YAML:\n%s", string(data))
		}
		path := filepath.Join(t.TempDir(), ".agenticlens-config.yaml")
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		reloaded, err := config.LoadFromFile(path)
		if err != nil {
			t.Fatalf("reloading generated config failed: %v", err)
		}
		if len(reloaded.ModelPricing.Rates) != 2 {
			t.Fatalf("reloaded rates count = %d, want 2", len(reloaded.ModelPricing.Rates))
		}
		if reloaded.ModelPricing.EffectiveCatalogID() != cfg.ModelPricing.EffectiveCatalogID() {
			t.Fatalf("catalog ID mismatch after round-trip: got %q want %q", reloaded.ModelPricing.EffectiveCatalogID(), cfg.ModelPricing.EffectiveCatalogID())
		}
	})
}
