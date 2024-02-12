package outbound_dnclist

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	gcloud "terraform-provider-genesyscloud/genesyscloud"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v121/platformclientv2"
)

func dataSourceOutboundDncListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sdkConfig := m.(*gcloud.ProviderMeta).ClientConfig
	outboundAPI := platformclientv2.NewOutboundApiWithConfig(sdkConfig)
	name := d.Get("name").(string)

	return gcloud.WithRetries(ctx, 15*time.Second, func() *retry.RetryError {
		const pageNum = 1
		const pageSize = 100
		dncLists, _, getErr := outboundAPI.GetOutboundDnclists(false, false, pageSize, pageNum, true, "", name, "", []string{}, "", "")
		if getErr != nil {
			return retry.NonRetryableError(fmt.Errorf("error requesting dnc lists %s: %s", name, getErr))
		}
		if dncLists.Entities == nil || len(*dncLists.Entities) == 0 {
			return retry.RetryableError(fmt.Errorf("no dnc lists found with name %s", name))
		}
		dncList := (*dncLists.Entities)[0]
		d.SetId(*dncList.Id)
		return nil
	})
}