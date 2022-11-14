// Copyright 2022 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package model

import (
	"github.com/pb33f/libopenapi/datamodel/low"
	v2 "github.com/pb33f/libopenapi/datamodel/low/v2"
	v3 "github.com/pb33f/libopenapi/datamodel/low/v3"
	"reflect"
)

type ComponentsChanges struct {
	PropertyChanges
	SchemaChanges         map[string]*SchemaChanges
	ResponsesChanges      map[string]*ResponseChanges
	ParameterChanges      map[string]*ParameterChanges
	ExamplesChanges       map[string]*ExampleChanges
	RequestBodyChanges    map[string]*RequestBodyChanges
	HeaderChanges         map[string]*HeaderChanges
	SecuritySchemeChanges map[string]*SecuritySchemeChanges
	LinkChanges           map[string]*LinkChanges
	CallbackChanges       map[string]*CallbackChanges
	ExtensionChanges      *ExtensionChanges
}

func CompareComponents(l, r any) *ComponentsChanges {

	var changes []*Change

	cc := new(ComponentsChanges)

	// Swagger Parameters
	if reflect.TypeOf(&v2.ParameterDefinitions{}) == reflect.TypeOf(l) &&
		reflect.TypeOf(&v2.ParameterDefinitions{}) == reflect.TypeOf(r) {
		lDef := l.(*v2.ParameterDefinitions)
		rDef := r.(*v2.ParameterDefinitions)
		cc.ParameterChanges = CheckMapForChanges(lDef.Definitions, rDef.Definitions, &changes,
			v3.ParametersLabel, CompareParametersV2)
	}

	// Swagger Responses
	if reflect.TypeOf(&v2.ResponsesDefinitions{}) == reflect.TypeOf(l) &&
		reflect.TypeOf(&v2.ResponsesDefinitions{}) == reflect.TypeOf(r) {
		lDef := l.(*v2.ResponsesDefinitions)
		rDef := r.(*v2.ResponsesDefinitions)
		cc.ResponsesChanges = CheckMapForChanges(lDef.Definitions, rDef.Definitions, &changes,
			v3.ResponsesLabel, CompareResponseV2)
	}

	// Swagger Schemas
	if reflect.TypeOf(&v2.Definitions{}) == reflect.TypeOf(l) &&
		reflect.TypeOf(&v2.Definitions{}) == reflect.TypeOf(r) {
		lDef := l.(*v2.Definitions)
		rDef := r.(*v2.Definitions)
		cc.SchemaChanges = CheckMapForChanges(lDef.Schemas, rDef.Schemas, &changes,
			v2.DefinitionsLabel, CompareSchemas)
	}

	// Swagger Security Definitions
	if reflect.TypeOf(&v2.SecurityDefinitions{}) == reflect.TypeOf(l) &&
		reflect.TypeOf(&v2.SecurityDefinitions{}) == reflect.TypeOf(r) {
		lDef := l.(*v2.SecurityDefinitions)
		rDef := r.(*v2.SecurityDefinitions)
		cc.SecuritySchemeChanges = CheckMapForChanges(lDef.Definitions, rDef.Definitions, &changes,
			v3.SecurityDefinitionLabel, CompareSecuritySchemesV2)
	}

	// OpenAPI Components
	if reflect.TypeOf(&v3.Components{}) == reflect.TypeOf(l) &&
		reflect.TypeOf(&v3.Components{}) == reflect.TypeOf(r) {

		lComponents := l.(*v3.Components)
		rComponents := r.(*v3.Components)

		doneChan := make(chan componentComparison)
		comparisons := 0

		// run as fast as we can, thread all the things.
		if !lComponents.Schemas.IsEmpty() || rComponents.Schemas.IsEmpty() {
			comparisons++
			go runComparison(lComponents.Schemas.Value, rComponents.Schemas.Value,
				&changes, v3.SchemasLabel, CompareSchemas, doneChan)
		}

		if !lComponents.Responses.IsEmpty() || rComponents.Responses.IsEmpty() {
			comparisons++
			go runComparison(lComponents.Responses.Value, rComponents.Responses.Value,
				&changes, v3.ResponsesLabel, CompareResponseV3, doneChan)
		}

		if !lComponents.Parameters.IsEmpty() || rComponents.Parameters.IsEmpty() {
			comparisons++
			go runComparison(lComponents.Parameters.Value, rComponents.Parameters.Value,
				&changes, v3.ParametersLabel, CompareParametersV3, doneChan)
		}

		if !lComponents.Examples.IsEmpty() || rComponents.Examples.IsEmpty() {
			comparisons++
			go runComparison(lComponents.Examples.Value, rComponents.Examples.Value,
				&changes, v3.ExamplesLabel, CompareExamples, doneChan)
		}

		if !lComponents.RequestBodies.IsEmpty() || rComponents.RequestBodies.IsEmpty() {
			comparisons++
			go runComparison(lComponents.RequestBodies.Value, rComponents.RequestBodies.Value,
				&changes, v3.RequestBodiesLabel, CompareRequestBodies, doneChan)
		}

		if !lComponents.Headers.IsEmpty() || rComponents.Headers.IsEmpty() {
			comparisons++
			go runComparison(lComponents.Headers.Value, rComponents.Headers.Value,
				&changes, v3.HeadersLabel, CompareHeadersV3, doneChan)
		}

		if !lComponents.SecuritySchemes.IsEmpty() || rComponents.SecuritySchemes.IsEmpty() {
			comparisons++
			go runComparison(lComponents.SecuritySchemes.Value, rComponents.SecuritySchemes.Value,
				&changes, v3.SecuritySchemesLabel, CompareSecuritySchemesV3, doneChan)
		}

		if !lComponents.Links.IsEmpty() || rComponents.Links.IsEmpty() {
			comparisons++
			go runComparison(lComponents.Links.Value, rComponents.Links.Value,
				&changes, v3.LinksLabel, CompareLinks, doneChan)
		}

		if !lComponents.Callbacks.IsEmpty() || rComponents.Callbacks.IsEmpty() {
			comparisons++
			go runComparison(lComponents.Callbacks.Value, rComponents.Callbacks.Value,
				&changes, v3.CallbacksLabel, CompareCallback, doneChan)
		}

		cc.ExtensionChanges = CompareExtensions(lComponents.Extensions, rComponents.Extensions)

		completedComponents := 0
		for completedComponents < comparisons {
			select {
			case res := <-doneChan:
				switch res.prop {
				case v3.SchemasLabel:
					completedComponents++
					cc.SchemaChanges = res.result.(map[string]*SchemaChanges)
					break
				case v3.ResponsesLabel:
					completedComponents++
					cc.ResponsesChanges = res.result.(map[string]*ResponseChanges)
					break
				case v3.ParametersLabel:
					completedComponents++
					cc.ParameterChanges = res.result.(map[string]*ParameterChanges)
					break
				case v3.ExamplesLabel:
					completedComponents++
					cc.ExamplesChanges = res.result.(map[string]*ExampleChanges)
					break
				case v3.RequestBodiesLabel:
					completedComponents++
					cc.RequestBodyChanges = res.result.(map[string]*RequestBodyChanges)
					break
				case v3.HeadersLabel:
					completedComponents++
					cc.HeaderChanges = res.result.(map[string]*HeaderChanges)
					break
				case v3.SecuritySchemesLabel:
					completedComponents++
					cc.SecuritySchemeChanges = res.result.(map[string]*SecuritySchemeChanges)
					break
				case v3.LinksLabel:
					completedComponents++
					cc.LinkChanges = res.result.(map[string]*LinkChanges)
					break
				case v3.CallbacksLabel:
					completedComponents++
					cc.CallbackChanges = res.result.(map[string]*CallbackChanges)
					break
				}
			}
		}
	}

	cc.Changes = changes
	if cc.TotalChanges() <= 0 {
		return nil
	}
	return cc
}

type componentComparison struct {
	prop   string
	result any
}

// run a generic comparison in a thread which in turn splits checks into further threads.
func runComparison[T any, R any](l, r map[low.KeyReference[string]]low.ValueReference[T],
	changes *[]*Change, label string, compareFunc func(l, r T) R, doneChan chan componentComparison) {
	doneChan <- componentComparison{
		prop:   label,
		result: CheckMapForChanges(l, r, changes, label, compareFunc),
	}
}

func (c *ComponentsChanges) TotalChanges() int {
	v := c.PropertyChanges.TotalChanges()
	for k := range c.SchemaChanges {
		v += c.SchemaChanges[k].TotalChanges()
	}
	for k := range c.ResponsesChanges {
		v += c.ResponsesChanges[k].TotalChanges()
	}
	for k := range c.ParameterChanges {
		v += c.ParameterChanges[k].TotalChanges()
	}
	for k := range c.ExamplesChanges {
		v += c.ExamplesChanges[k].TotalChanges()
	}
	for k := range c.RequestBodyChanges {
		v += c.RequestBodyChanges[k].TotalChanges()
	}
	for k := range c.HeaderChanges {
		v += c.HeaderChanges[k].TotalChanges()
	}
	for k := range c.SecuritySchemeChanges {
		v += c.SecuritySchemeChanges[k].TotalChanges()
	}
	for k := range c.LinkChanges {
		v += c.LinkChanges[k].TotalChanges()
	}
	for k := range c.CallbackChanges {
		v += c.CallbackChanges[k].TotalChanges()
	}
	if c.ExtensionChanges != nil {
		v += c.ExtensionChanges.TotalChanges()
	}
	return v
}

func (c *ComponentsChanges) TotalBreakingChanges() int {
	v := c.PropertyChanges.TotalBreakingChanges()
	for k := range c.SchemaChanges {
		v += c.SchemaChanges[k].TotalBreakingChanges()
	}
	for k := range c.ResponsesChanges {
		v += c.ResponsesChanges[k].TotalBreakingChanges()
	}
	for k := range c.ParameterChanges {
		v += c.ParameterChanges[k].TotalBreakingChanges()
	}
	for k := range c.ExamplesChanges {
		v += c.ExamplesChanges[k].TotalBreakingChanges()
	}
	for k := range c.RequestBodyChanges {
		v += c.RequestBodyChanges[k].TotalBreakingChanges()
	}
	for k := range c.HeaderChanges {
		v += c.HeaderChanges[k].TotalBreakingChanges()
	}
	for k := range c.SecuritySchemeChanges {
		v += c.SecuritySchemeChanges[k].TotalBreakingChanges()
	}
	for k := range c.LinkChanges {
		v += c.LinkChanges[k].TotalBreakingChanges()
	}
	for k := range c.CallbackChanges {
		v += c.CallbackChanges[k].TotalBreakingChanges()
	}
	if c.ExtensionChanges != nil {
		v += c.ExtensionChanges.TotalBreakingChanges()
	}
	return v
}