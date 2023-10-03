package integration_custom_auth_action

import (
	"sync"
	"testing"

	integration "terraform-provider-genesyscloud/genesyscloud/integration"
	integrationCred "terraform-provider-genesyscloud/genesyscloud/integration_credential"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
   The genesyscloud_integration_action_init_test.go file is used to initialize the data sources and resources
   used in testing the integration action resource.

   Please make sure you register ALL resources and data sources your test cases will use.
*/

const (
	trueValue  = "true"
	falseValue = "false"
	nullValue  = "null"
)

// providerDataSources holds a map of all registered datasources
var providerDataSources map[string]*schema.Resource

// providerResources holds a map of all registered resources
var providerResources map[string]*schema.Resource

type registerTestInstance struct {
	resourceMapMutex sync.RWMutex
}

// registerTestResources registers all resources used in the tests
func (r *registerTestInstance) registerTestResources() {
	r.resourceMapMutex.Lock()
	defer r.resourceMapMutex.Unlock()

	providerResources[resourceName] = ResourceIntegrationCustomAuthAction()
	providerResources["genesyscloud_integration"] = integration.ResourceIntegration()
	providerResources["genesyscloud_integration_credential"] = integrationCred.ResourceIntegrationCredential()
}

// initTestresources initializes all test resources and data sources.
func initTestresources() {
	providerDataSources = make(map[string]*schema.Resource)
	providerResources = make(map[string]*schema.Resource)

	reg_instance := &registerTestInstance{}

	reg_instance.registerTestResources()
}

// TestMain is a "setup" function called by the testing framework when run the test
func TestMain(m *testing.M) {
	// Run setup function before starting the test suite for integration package
	initTestresources()

	// Run the test suite for suite for the integration package
	m.Run()
}
