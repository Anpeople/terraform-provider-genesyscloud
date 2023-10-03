package integration_custom_auth_action

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/util/resourcedata"

	gcloud "terraform-provider-genesyscloud/genesyscloud"
	integrationAction "terraform-provider-genesyscloud/genesyscloud/integration_action"
	resourceExporter "terraform-provider-genesyscloud/genesyscloud/resource_exporter"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v109/platformclientv2"
)

/*
The resource_genesyscloud_integration_action.go contains all of the methods that perform the core logic for a resource.
In general a resource should have a approximately 5 methods in it:

1.  A getAll.... function that the CX as Code exporter will use during the process of exporting Genesys Cloud.
2.  A create.... function that the resource will use to create a Genesys Cloud object (e.g. genesycloud_integration_action)
3.  A read.... function that looks up a single resource.
4.  An update... function that updates a single resource.
5.  A delete.... function that deletes a single resource.

Two things to note:

 1. All code in these methods should be focused on getting data in and out of Terraform.  All code that is used for interacting
    with a Genesys API should be encapsulated into a proxy class contained within the package.

 2. In general, to keep this file somewhat manageable, if you find yourself with a number of helper functions move them to a

utils function in the package.  This will keep the code manageable and easy to work through.
*/

// getAllIntegrationActions retrieves all of the integration action via Terraform in the Genesys Cloud and is used for the exporter
func getAllIntegrationCustomAuthActions(ctx context.Context, clientConfig *platformclientv2.Configuration) (resourceExporter.ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(resourceExporter.ResourceIDMetaMap)
	cap := getCustomAuthActionsProxy(clientConfig)

	actions, err := cap.getAllIntegrationCustomAuthActions(ctx)
	if err != nil {
		return nil, diag.Errorf("Failed to get integration custom auth actions: %v", err)
	}

	for _, action := range *actions {
		resources[*action.Id] = &resourceExporter.ResourceMeta{Name: *action.Name}
	}

	return resources, nil
}

// createIntegrationAction is used by the integration actions resource to create Genesyscloud integration action
func createIntegrationCustomAuthAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	cap := getCustomAuthActionsProxy(sdkConfig)

	integrationId := d.Get("integration_id").(string)
	authActionId := getCustomAuthIdFromIntegration(integrationId)

	name := resourcedata.GetNillableValue[string](d, "name")

	// Precheck that integration type and its credential type if it should have a custom auth data action
	if ok, err := isIntegrationAndCredTypesCorrect(ctx, cap, integrationId); !ok || err != nil {
		return diag.Errorf("configuration of integration %s does not allow for a custom auth data action. %v", integrationId, err)
	}

	log.Printf("Retrieving the custom auth action of integration %s", integrationId)

	// Retrieve the automatically-generated custom auth action
	// to set the resource Id
	diagErr := gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		authAction, resp, err := cap.getCustomAuthActionById(ctx, authActionId)
		if err != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("cannot find custom auth action of integration %s: %v", integrationId, err))
			}
			return retry.NonRetryableError(fmt.Errorf("error deleting integration %s: %s", d.Id(), err))
		}

		// Get default name if not to be overriden
		if name == nil {
			name = authAction.Name
		}

		d.SetId(*authAction.Id)

		return nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updating custom auth action of integration %s", integrationId)

	// Update the custom auth action with the actual configuration
	diagErr = gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get the latest action version to send with PATCH
		action, resp, err := cap.getCustomAuthActionById(ctx, authActionId)
		if err != nil {
			return resp, diag.Errorf("Failed to read integration custom auth action %s: %s", authActionId, err)
		}

		_, resp, err = cap.updateCustomAuthAction(ctx, authActionId, &platformclientv2.Updateactioninput{
			Name:    name,
			Version: action.Version,
			Config:  BuildSdkCustomAuthActionConfig(d),
		})
		if err != nil {
			return resp, diag.Errorf("Failed to update integration action %s: %s", *name, err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated custom auth action %s", *name)

	return readIntegrationCustomAuthAction(ctx, d, meta)
}

// readIntegrationAction is used by the integration action resource to read an action from genesys cloud.
func readIntegrationCustomAuthAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	cap := getCustomAuthActionsProxy(sdkConfig)

	log.Printf("Reading integration action %s", d.Id())

	return gcloud.WithRetriesForRead(ctx, d, func() *retry.RetryError {
		action, resp, err := cap.getCustomAuthActionById(ctx, d.Id())
		if err != nil {
			if gcloud.IsStatus404(resp) {
				return retry.RetryableError(fmt.Errorf("failed to read integration custom auth action %s: %s", d.Id(), err))
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read integration custom auth action %s: %s", d.Id(), err))
		}

		// Retrieve config request/response templates
		reqTemp, resp, err := cap.getIntegrationActionTemplate(ctx, d.Id(), reqTemplateFileName)
		if err != nil {
			if gcloud.IsStatus404(resp) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read request template for integration action %s: %s", d.Id(), err))
		}

		successTemp, resp, err := cap.getIntegrationActionTemplate(ctx, d.Id(), successTemplateFileName)
		if err != nil {
			if gcloud.IsStatus404(resp) {
				d.SetId("")
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("failed to read success template for integration action %s: %s", d.Id(), err))
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, ResourceIntegrationCustomAuthAction())

		resourcedata.SetNillableValue(d, "name", action.Name)
		resourcedata.SetNillableValue(d, "integration_id", action.IntegrationId)

		if action.Config != nil && action.Config.Request != nil {
			action.Config.Request.RequestTemplate = reqTemp
			d.Set("config_request", integrationAction.FlattenActionConfigRequest(*action.Config.Request))
		} else {
			d.Set("config_request", nil)
		}

		if action.Config != nil && action.Config.Response != nil {
			action.Config.Response.SuccessTemplate = successTemp
			d.Set("config_response", integrationAction.FlattenActionConfigResponse(*action.Config.Response))
		} else {
			d.Set("config_response", nil)
		}

		log.Printf("Read integration action %s %s", d.Id(), *action.Name)
		return cc.CheckState()
	})
}

// updateIntegrationAction is used by the integration action resource to update an action in Genesys Cloud
func updateIntegrationCustomAuthAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*gcloud.ProviderMeta).ClientConfig
	cap := getCustomAuthActionsProxy(sdkConfig)

	name := resourcedata.GetNillableValue[string](d, "name")

	log.Printf("Updating integration custom auth action %s", *name)

	diagErr := gcloud.RetryWhen(gcloud.IsVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get the latest action version to send with PATCH
		action, resp, err := cap.getCustomAuthActionById(ctx, d.Id())
		if err != nil {
			return resp, diag.Errorf("Failed to read integration custom auth action %s: %s", d.Id(), err)
		}
		if name == nil {
			name = action.Name
		}

		_, resp, err = cap.updateCustomAuthAction(ctx, d.Id(), &platformclientv2.Updateactioninput{
			Name:    name,
			Version: action.Version,
			Config:  BuildSdkCustomAuthActionConfig(d),
		})
		if err != nil {
			return resp, diag.Errorf("Failed to update integration action %s: %s", *name, err)
		}
		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated custom auth action %s", *name)

	return readIntegrationCustomAuthAction(ctx, d, meta)
}

// deleteIntegrationAction is used by the integration action resource to delete an action from Genesys cloud.
func deleteIntegrationCustomAuthAction(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	log.Printf("Removing terraform resource integration_custom_auth_action %s will not remove the Data Action itself in the org", name)
	log.Printf("The Custom Auth Data Action cannot be removed unless the Web Services Data Action Integration itself is deleted or if the Credentials type is changed from 'User Defined (OAuth)' to a different type")

	return nil
}
