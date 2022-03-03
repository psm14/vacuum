// Copyright 2020-2022 Dave Shanley / Quobix
// SPDX-License-Identifier: MIT

package rulesets

import (
	"github.com/daveshanley/vacuum/model"
	"github.com/daveshanley/vacuum/parser"
	"github.com/daveshanley/vacuum/utils"
	"sync"
)

const (
	warn       = "warn"
	error      = "error"
	info       = "info"
	hint       = "hint"
	style      = "style"
	validation = "validation"
)

type ruleSetsModel struct {
	openAPIRuleSet *model.RuleSet
}

// RuleSets is used to generate default RuleSets built into vacuum
type RuleSets interface {

	// GenerateOpenAPIDefaultRuleSet generates a ready to run pointer to a model.RuleSet containing all
	// OpenAPI rules supported by vacuum. Passing all these rules would be considered a good quality specification.
	GenerateOpenAPIDefaultRuleSet() *model.RuleSet
}

var rulesetsSingleton *ruleSetsModel
var openAPIRulesGrab sync.Once

func BuildDefaultRuleSets() RuleSets {
	openAPIRulesGrab.Do(func() {
		rulesetsSingleton = &ruleSetsModel{
			openAPIRuleSet: generateDefaultOpenAPIRuleSet(),
		}
	})

	return rulesetsSingleton
}

func (rsm ruleSetsModel) GenerateOpenAPIDefaultRuleSet() *model.RuleSet {
	return rsm.openAPIRuleSet
}

func generateDefaultOpenAPIRuleSet() *model.RuleSet {

	rules := make(map[string]*model.Rule)

	// add success response
	rules["operation-success-response"] = &model.Rule{
		Description: "Operation must have at least one 2xx or a 3xx response.",
		Given:       "$",
		Resolved:    true,
		Recommended: true,
		Type:        style,
		Severity:    warn,
		Then: model.RuleAction{
			Field:    "responses",
			Function: "oasOpSuccessResponse",
		},
	}

	// add unique operation ID rule
	rules["operation-operationId-unique"] = &model.Rule{
		Description: "Every operation must have unique \"operationId\".",
		Given:       "$.paths",
		Resolved:    true,
		Recommended: true,
		Type:        validation,
		Severity:    warn,
		Then: model.RuleAction{
			Function: "oasOpIdUnique",
		},
	}

	// add operation params rule
	rules["operation-parameters"] = &model.Rule{
		Description: "Operation parameters are unique and non-repeating.",
		Given:       "$.paths",
		Resolved:    true,
		Recommended: true,
		Type:        validation,
		Severity:    warn,
		Then: model.RuleAction{
			Function: "oasOpParams",
		},
	}

	// add operation tag defined rule
	rules["operation-tag-defined"] = &model.Rule{
		Description: "Operation tags must be defined in global tags.",
		Given:       "$",
		Resolved:    true,
		Recommended: true,
		Type:        validation,
		Severity:    warn,
		Then: model.RuleAction{
			Function: "oasTagDefined",
		},
	}

	// add operation tag defined rule
	rules["path-params"] = &model.Rule{
		Description: "Path parameters must be defined and valid.",
		Given:       "$",
		Resolved:    true,
		Recommended: true,
		Type:        validation,
		Severity:    error,
		Then: model.RuleAction{
			Function: "oasPathParam",
		},
	}

	// contact-properties
	rules["contact-properties"] = GetContactPropertiesRule()

	// info object: contains contact
	rules["info-contact"] = GetInfoContactRule()

	// info object: contains description
	rules["info-description"] = GetInfoDescriptionRule()

	// info object: contains a license
	rules["info-license"] = GetInfoLicenseRule()

	// info object: contains a license url
	rules["license-url"] = GetInfoLicenseUrlRule()

	// duplicated entry in enums
	duplicatedEnum := make(map[string]interface{})
	duplicatedEnum["schema"] = parser.Schema{
		Type: &utils.ArrayLabel,
		Items: &parser.Schema{
			Type: &utils.StringLabel,
		},
		UniqueItems: true,
	}

	rules["duplicated-entry-in-enum"] = &model.Rule{
		Description: "Enum values must not have duplicate entry",
		Given:       "$..[?(@.enum)]",
		Resolved:    true,
		Recommended: true,
		Type:        validation,
		Severity:    error,
		Then: model.RuleAction{
			Field:           "enum",
			Function:        "oasSchema",
			FunctionOptions: duplicatedEnum,
		},
	}

	// add no $ref siblings
	rules["no-$ref-siblings"] = &model.Rule{
		Description: "$ref values cannot be placed next to other properties (like a description)",
		Given:       "$",
		Resolved:    false,
		Recommended: true,
		Type:        validation,
		Severity:    error,
		Then: model.RuleAction{
			Function: "refSiblings",
		},
	}

	// add unused component rule for openapi3
	unusedComponentRule := &model.Rule{
		Description: "Check for unused components and bad references",
		Given:       "$",
		Resolved:    false,
		Recommended: true,
		Type:        validation,
		Severity:    warn,
		Then: model.RuleAction{
			Function: "oasUnusedComponent",
		},
	}

	rules["oas3-unused-component"] = unusedComponentRule
	// TODO: build in spec types so we don't run this twice :)
	//rules["oas2-unused-definition"] = unusedComponentRule

	// swagger operation security values defined
	oasSecurityPath := make(map[string]string)
	oasSecurityPath["schemesPath"] = "$.components.securitySchemes"

	rules["oas3-operation-security-defined"] = &model.Rule{
		Description: "'security' values must match a scheme defined in components.securitySchemes",
		Given:       "$",
		Resolved:    true,
		Recommended: true,
		Type:        validation,
		Severity:    error,
		Then: model.RuleAction{
			Function:        "oasOpSecurityDefined",
			FunctionOptions: oasSecurityPath,
		},
	}

	// swagger operation security values defined
	swaggerSecurityPath := make(map[string]string)
	swaggerSecurityPath["schemesPath"] = "$.securityDefinitions"

	rules["oas2-operation-security-defined"] = &model.Rule{
		Description: "'security' values must match a scheme defined in securityDefinitions",
		Given:       "$",
		Resolved:    true,
		Recommended: true,
		Type:        validation,
		Severity:    error,
		Then: model.RuleAction{
			Function:        "oasOpSecurityDefined",
			FunctionOptions: swaggerSecurityPath,
		},
	}

	// check all examples
	rules["oas-3valid-schema-example"] = &model.Rule{
		Description: "Examples must be present",
		Given:       "$",
		Resolved:    true,
		Recommended: true,
		Type:        validation,
		Severity:    error,
		Then: model.RuleAction{
			Function: "oasExample",
		},
	}

	set := &model.RuleSet{
		DocumentationURI: "https://quobix.com/vacuum/rules/openapi",
		Rules:            rules,
	}

	return set

}
