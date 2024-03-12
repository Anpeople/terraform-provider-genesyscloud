package routing_email_route

import (
	"context"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v123/platformclientv2"
	"log"
)

/*
The genesyscloud_routing_email_route_proxy.go file contains the proxy structures and methods that interact
with the Genesys Cloud SDK. We use composition here for each function on the proxy so individual functions can be stubbed
out during testing.
*/

// internalProxy holds a proxy instance that can be used throughout the package
var internalProxy *routingEmailRouteProxy

// Type definitions for each func on our proxy so we can easily mock them out later
type createRoutingEmailRouteFunc func(ctx context.Context, p *routingEmailRouteProxy, domainId string, inboundRoute *platformclientv2.Inboundroute) (*platformclientv2.Inboundroute, int, error)
type getAllRoutingEmailRouteFunc func(ctx context.Context, p *routingEmailRouteProxy, domainId string, name string) (*[]platformclientv2.Inboundroute, int, error)
type getRoutingEmailRouteIdByNameFunc func(ctx context.Context, p *routingEmailRouteProxy, name string) (id string, retryable bool, respCode int, err error)
type getRoutingEmailRouteByIdFunc func(ctx context.Context, p *routingEmailRouteProxy, domainId string, id string) (inboundRoute *platformclientv2.Inboundroute, responseCode int, err error)
type updateRoutingEmailRouteFunc func(ctx context.Context, p *routingEmailRouteProxy, id string, domainId string, inboundRoute *platformclientv2.Inboundroute) (*platformclientv2.Inboundroute, int, error)
type deleteRoutingEmailRouteFunc func(ctx context.Context, p *routingEmailRouteProxy, domainId string, id string) (responseCode int, err error)

// routingEmailRouteProxy contains all of the methods that call genesys cloud APIs.
type routingEmailRouteProxy struct {
	clientConfig                     *platformclientv2.Configuration
	routingApi                       *platformclientv2.RoutingApi
	createRoutingEmailRouteAttr      createRoutingEmailRouteFunc
	getAllRoutingEmailRouteAttr      getAllRoutingEmailRouteFunc
	getRoutingEmailRouteIdByNameAttr getRoutingEmailRouteIdByNameFunc
	getRoutingEmailRouteByIdAttr     getRoutingEmailRouteByIdFunc
	updateRoutingEmailRouteAttr      updateRoutingEmailRouteFunc
	deleteRoutingEmailRouteAttr      deleteRoutingEmailRouteFunc
}

// newRoutingEmailRouteProxy initializes the routing email route proxy with all of the data needed to communicate with Genesys Cloud
func newRoutingEmailRouteProxy(clientConfig *platformclientv2.Configuration) *routingEmailRouteProxy {
	api := platformclientv2.NewRoutingApiWithConfig(clientConfig)
	return &routingEmailRouteProxy{
		clientConfig:                     clientConfig,
		routingApi:                       api,
		createRoutingEmailRouteAttr:      createRoutingEmailRouteFn,
		getAllRoutingEmailRouteAttr:      getAllRoutingEmailRouteFn,
		getRoutingEmailRouteIdByNameAttr: getRoutingEmailRouteIdByNameFn,
		getRoutingEmailRouteByIdAttr:     getRoutingEmailRouteByIdFn,
		updateRoutingEmailRouteAttr:      updateRoutingEmailRouteFn,
		deleteRoutingEmailRouteAttr:      deleteRoutingEmailRouteFn,
	}
}

// getRoutingEmailRouteProxy acts as a singleton to for the internalProxy.  It also ensures
// that we can still proxy our tests by directly setting internalProxy package variable
func getRoutingEmailRouteProxy(clientConfig *platformclientv2.Configuration) *routingEmailRouteProxy {
	if internalProxy == nil {
		internalProxy = newRoutingEmailRouteProxy(clientConfig)
	}
	return internalProxy
}

// createRoutingEmailRoute creates a Genesys Cloud routing email route
func (p *routingEmailRouteProxy) createRoutingEmailRoute(ctx context.Context, domainId string, routingEmailRoute *platformclientv2.Inboundroute) (*platformclientv2.Inboundroute, int, error) {
	return p.createRoutingEmailRouteAttr(ctx, p, domainId, routingEmailRoute)
}

// getRoutingEmailRoute retrieves all Genesys Cloud routing email route
func (p *routingEmailRouteProxy) getAllRoutingEmailRoute(ctx context.Context, domainId string, name string) (*[]platformclientv2.Inboundroute, int, error) {
	return p.getAllRoutingEmailRouteAttr(ctx, p, domainId, name)
}

// getRoutingEmailRouteIdByName returns a single Genesys Cloud routing email route by a name
func (p *routingEmailRouteProxy) getRoutingEmailRouteIdByName(ctx context.Context, name string) (id string, retryable bool, respCode int, err error) {
	return p.getRoutingEmailRouteIdByNameAttr(ctx, p, name)
}

// getRoutingEmailRouteById returns a single Genesys Cloud routing email route by Id
func (p *routingEmailRouteProxy) getRoutingEmailRouteById(ctx context.Context, domainId string, id string) (routingEmailRoute *platformclientv2.Inboundroute, statusCode int, err error) {
	return p.getRoutingEmailRouteByIdAttr(ctx, p, domainId, id)
}

// updateRoutingEmailRoute updates a Genesys Cloud routing email route
func (p *routingEmailRouteProxy) updateRoutingEmailRoute(ctx context.Context, id string, domainId string, routingEmailRoute *platformclientv2.Inboundroute) (*platformclientv2.Inboundroute, int, error) {
	return p.updateRoutingEmailRouteAttr(ctx, p, id, domainId, routingEmailRoute)
}

// deleteRoutingEmailRoute deletes a Genesys Cloud routing email route by Id
func (p *routingEmailRouteProxy) deleteRoutingEmailRoute(ctx context.Context, domainId string, id string) (statusCode int, err error) {
	return p.deleteRoutingEmailRouteAttr(ctx, p, domainId, id)
}

// getAllRoutingEmailRouteFn is the implementation for retrieving all routing email route in Genesys Cloud
func getAllRoutingEmailRouteFn(ctx context.Context, p *routingEmailRouteProxy, domainId string, name string) (*[]platformclientv2.Inboundroute, int, error) {
	var allInboundRoutes []platformclientv2.Inboundroute
	const pageSize = 100
	var statusCode int

	// If domainID is given, we only return the routes for that specific domain
	if domainId != "" {
		routes, resp, err := p.routingApi.GetRoutingEmailDomainRoutes(domainId, pageSize, 1, name)
		if err != nil {
			return nil, 0, fmt.Errorf("Failed to get routing email route: %s", err)
		}
		if routes.Entities == nil || len(*routes.Entities) == 0 {
			return &allInboundRoutes, resp.StatusCode, nil
		}

		for _, route := range *routes.Entities {
			allInboundRoutes = append(allInboundRoutes, route)
		}
		return &allInboundRoutes, resp.StatusCode, nil
	}

	// DomainID not given so we must acquire every route for every domain

	domains, resp, err := p.routingApi.GetRoutingEmailDomains(pageSize, 1, false, "")
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to get routing email domains: %s", err)
	}
	if domains.Entities == nil || len(*domains.Entities) == 0 {
		return &allInboundRoutes, resp.StatusCode, nil
	}

	for _, domain := range *domains.Entities {
		for pageNum := 1; ; pageNum++ {
			routes, _, err := p.routingApi.GetRoutingEmailDomainRoutes(*domain.Id, pageSize, pageNum, name)
			if err != nil {
				return nil, 0, fmt.Errorf("Failed to get routing email route: %s", err)
			}

			if routes.Entities == nil || len(*routes.Entities) == 0 {
				break
			}

			for _, route := range *routes.Entities {
				allInboundRoutes = append(allInboundRoutes, route)
			}
		}
	}

	for pageNum := 2; pageNum <= *domains.PageCount; pageNum++ {
		domains, resp, err := p.routingApi.GetRoutingEmailDomains(pageSize, pageNum, false, "")
		if err != nil {
			return nil, 0, fmt.Errorf("Failed to get routing email domains: %s", err)
		}

		if domains.Entities == nil || len(*domains.Entities) == 0 {
			return &allInboundRoutes, resp.StatusCode, nil
		}

		for _, domain := range *domains.Entities {
			for pageNum := 1; ; pageNum++ {
				routes, _, err := p.routingApi.GetRoutingEmailDomainRoutes(*domain.Id, pageSize, pageNum, name)
				if err != nil {
					return nil, 0, fmt.Errorf("Failed to get routing email route: %s", err)
				}

				if routes.Entities == nil || len(*routes.Entities) == 0 {
					break
				}

				for _, route := range *routes.Entities {
					allInboundRoutes = append(allInboundRoutes, route)
				}
			}
		}
	}
	return &allInboundRoutes, statusCode, nil
}

// createRoutingEmailRouteFn is an implementation function for creating a Genesys Cloud routing email route
func createRoutingEmailRouteFn(ctx context.Context, p *routingEmailRouteProxy, domainId string, routingEmailRoute *platformclientv2.Inboundroute) (*platformclientv2.Inboundroute, int, error) {
	inboundRoute, resp, err := p.routingApi.PostRoutingEmailDomainRoutes(domainId, *routingEmailRoute)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed to create routing email route: %s", err)
	}
	return inboundRoute, resp.StatusCode, nil
}

// updateRoutingEmailRouteFn is an implementation of the function to update a Genesys Cloud routing email route
func updateRoutingEmailRouteFn(ctx context.Context, p *routingEmailRouteProxy, id string, domainId string, routingEmailRoute *platformclientv2.Inboundroute) (*platformclientv2.Inboundroute, int, error) {
	inboundRoute, resp, err := p.routingApi.PutRoutingEmailDomainRoute(domainId, id, *routingEmailRoute)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to update routing email route: %s", err)
	}
	return inboundRoute, resp.StatusCode, nil
}

// deleteRoutingEmailRouteFn is an implementation function for deleting a Genesys Cloud routing email route
func deleteRoutingEmailRouteFn(ctx context.Context, p *routingEmailRouteProxy, domainId string, id string) (statusCode int, err error) {
	resp, err := p.routingApi.DeleteRoutingEmailDomainRoute(domainId, id)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("Failed to delete routing email route: %s", err)
	}
	return resp.StatusCode, nil
}

// getRoutingEmailRouteByIdFn is an implementation of the function to get a Genesys Cloud routing email route by Id
func getRoutingEmailRouteByIdFn(ctx context.Context, p *routingEmailRouteProxy, domainId string, id string) (routingEmailRoute *platformclientv2.Inboundroute, statusCode int, err error) {
	inboundRoute, resp, err := p.routingApi.GetRoutingEmailDomainRoute(domainId, id)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("Failed to retrieve routing email route by id %s: %s", id, err)
	}

	return inboundRoute, resp.StatusCode, nil
}

// getRoutingEmailRouteIdByNameFn is an implementation of the function to get a Genesys Cloud routing email route by name
func getRoutingEmailRouteIdByNameFn(ctx context.Context, p *routingEmailRouteProxy, name string) (id string, retryable bool, respCode int, err error) {
	inboundRoutes, resp, err := getAllRoutingEmailRouteFn(ctx, p, "", name)
	if err != nil {
		return "", false, resp, err
	}

	if inboundRoutes == nil || len(*inboundRoutes) == 0 {
		return "", true, resp, fmt.Errorf("No routing email route found with name %s", name)
	}

	for _, inboundRoute := range *inboundRoutes {
		if *inboundRoute.Name == name {
			log.Printf("Retrieved the routing email route id %s by name %s", *inboundRoute.Id, name)
			return *inboundRoute.Id, false, resp, nil
		}
	}

	return "", true, resp, fmt.Errorf("Unable to find routing email route with name %s", name)
}