package handlers

import (
	"context"
	"errors"

	"code.cloudfoundry.org/copilot/api"
	"code.cloudfoundry.org/lager"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CAPI struct {
	Logger                           lager.Logger
	RoutesRepo                       routesRepoInterface
	RouteMappingsRepo                routeMappingsRepoInterface
	CAPIDiegoProcessAssociationsRepo capiDiegoProcessAssociationsRepoInterface
}

func (c *CAPI) Health(context.Context, *api.HealthRequest) (*api.HealthResponse, error) {
	c.Logger.Info("capi health check...")
	return &api.HealthResponse{Healthy: true}, nil
}

// TODO: probably remove or test these eventually, currently using for debugging
func (c *CAPI) ListCfRoutes(context.Context, *api.ListCfRoutesRequest) (*api.ListCfRoutesResponse, error) {
	c.Logger.Info("listing cf routes...")
	return &api.ListCfRoutesResponse{Routes: c.RoutesRepo.List()}, nil
}

// TODO: probably remove or test these eventually, currently using for debugging
func (c *CAPI) ListCfRouteMappings(context.Context, *api.ListCfRouteMappingsRequest) (*api.ListCfRouteMappingsResponse, error) {
	c.Logger.Info("listing cf route mappings...")
	routeMappings := c.RouteMappingsRepo.List()
	apiRoutMappings := make(map[string]*api.RouteMapping)
	for k, v := range routeMappings {
		apiRoutMappings[k] = &api.RouteMapping{
			CapiProcessGuid: string(v.CAPIProcessGUID),
			RouteGuid:       string(v.RouteGUID),
		}
	}
	return &api.ListCfRouteMappingsResponse{RouteMappings: apiRoutMappings}, nil
}

// TODO: probably remove or test these eventually, currently using for debugging
func (c *CAPI) ListCapiDiegoProcessAssociations(context.Context, *api.ListCapiDiegoProcessAssociationsRequest) (*api.ListCapiDiegoProcessAssociationsResponse, error) {
	c.Logger.Info("listing capi/diego process associations...")

	response := &api.ListCapiDiegoProcessAssociationsResponse{
		CapiDiegoProcessAssociations: make(map[string]*api.DiegoProcessGuids),
	}
	for capiProcessGUID, diegoProcessGUIDs := range c.CAPIDiegoProcessAssociationsRepo.List() {
		response.CapiDiegoProcessAssociations[string(capiProcessGUID)] = &api.DiegoProcessGuids{diegoProcessGUIDs.ToStringSlice()}
	}
	return response, nil
}

func (c *CAPI) UpsertRoute(context context.Context, request *api.UpsertRouteRequest) (*api.UpsertRouteResponse, error) {
	c.Logger.Info("upserting route...")
	err := validateUpsertRouteRequest(request)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Route %#v is invalid:\n %v", request, err)
	}
	route := &Route{
		GUID: RouteGUID(request.Route.Guid),
		Host: request.Route.Host,
	}
	c.RoutesRepo.Upsert(route)
	return &api.UpsertRouteResponse{}, nil
}

func (c *CAPI) DeleteRoute(context context.Context, request *api.DeleteRouteRequest) (*api.DeleteRouteResponse, error) {
	c.Logger.Info("deleting route...")
	err := validateDeleteRouteRequest(request)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err)
	}
	c.RoutesRepo.Delete(RouteGUID(request.Guid))
	return &api.DeleteRouteResponse{}, nil
}

func (c *CAPI) MapRoute(context context.Context, request *api.MapRouteRequest) (*api.MapRouteResponse, error) {
	c.Logger.Info("mapping route...")
	err := validateMapRouteRequest(request)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Route Mapping %#v is invalid:\n %v", request, err)
	}
	r := RouteMapping{
		RouteGUID:       RouteGUID(request.RouteMapping.RouteGuid),
		CAPIProcessGUID: CAPIProcessGUID(request.RouteMapping.CapiProcessGuid),
	}
	c.RouteMappingsRepo.Map(r)
	return &api.MapRouteResponse{}, nil
}

func (c *CAPI) UnmapRoute(context context.Context, request *api.UnmapRouteRequest) (*api.UnmapRouteResponse, error) {
	c.Logger.Info("unmapping route...")
	err := validateUnmapRouteRequest(request)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Route Mapping %#v is invalid:\n %v", request, err)
	}
	r := RouteMapping{RouteGUID: RouteGUID(request.RouteMapping.RouteGuid), CAPIProcessGUID: CAPIProcessGUID(request.RouteMapping.CapiProcessGuid)}
	c.RouteMappingsRepo.Unmap(r)
	return &api.UnmapRouteResponse{}, nil
}

func (c *CAPI) UpsertCapiDiegoProcessAssociation(context context.Context, request *api.UpsertCapiDiegoProcessAssociationRequest) (*api.UpsertCapiDiegoProcessAssociationResponse, error) {
	c.Logger.Info("upserting capi/diego process association...")
	err := validateUpsertCAPIDiegoProcessAssociationRequest(request)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Capi/Diego Process Association %#v is invalid:\n %v", request, err)
	}
	association := CAPIDiegoProcessAssociation{
		CAPIProcessGUID:   CAPIProcessGUID(request.CapiDiegoProcessAssociation.CapiProcessGuid),
		DiegoProcessGUIDs: DiegoProcessGUIDsFromStringSlice(request.CapiDiegoProcessAssociation.DiegoProcessGuids),
	}
	c.CAPIDiegoProcessAssociationsRepo.Upsert(association)
	return &api.UpsertCapiDiegoProcessAssociationResponse{}, nil
}

func (c *CAPI) DeleteCapiDiegoProcessAssociation(context context.Context, request *api.DeleteCapiDiegoProcessAssociationRequest) (*api.DeleteCapiDiegoProcessAssociationResponse, error) {
	c.Logger.Info("deleting capi/diego process association...")
	err := validateDeleteCAPIDiegoProcessAssociationRequest(request)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err)
	}

	c.CAPIDiegoProcessAssociationsRepo.Delete(CAPIProcessGUID(request.CapiProcessGuid))

	return &api.DeleteCapiDiegoProcessAssociationResponse{}, nil
}

func validateUpsertRouteRequest(r *api.UpsertRouteRequest) error {
	route := r.Route
	if route == nil {
		return errors.New("route is required")
	}
	if route.Guid == "" || route.Host == "" {
		return errors.New("route Guid and Host are required")
	}
	return nil
}

func validateDeleteRouteRequest(r *api.DeleteRouteRequest) error {
	if r.Guid == "" {
		return errors.New("route Guid is required")
	}
	return nil
}

func validateMapRouteRequest(r *api.MapRouteRequest) error {
	rm := r.RouteMapping
	if rm == nil {
		return errors.New("RouteMapping is required")
	}
	if rm.RouteGuid == "" || rm.CapiProcessGuid == "" {
		return errors.New("RouteGUID and CapiProcessGUID are required")
	}
	return nil
}

func validateUnmapRouteRequest(r *api.UnmapRouteRequest) error {
	rm := r.RouteMapping
	if rm == nil {
		return errors.New("RouteMapping is required")
	}
	if rm.RouteGuid == "" || rm.CapiProcessGuid == "" {
		return errors.New("RouteGuid and CapiProcessGuid are required")
	}
	return nil
}

func validateUpsertCAPIDiegoProcessAssociationRequest(r *api.UpsertCapiDiegoProcessAssociationRequest) error {
	association := r.CapiDiegoProcessAssociation
	if association == nil {
		return errors.New("CapiDiegoProcessAssociation is required")
	}
	if association.CapiProcessGuid == "" || len(association.DiegoProcessGuids) == 0 {
		return errors.New("CapiProcessGuid and DiegoProcessGuids are required")
	}
	return nil
}

func validateDeleteCAPIDiegoProcessAssociationRequest(r *api.DeleteCapiDiegoProcessAssociationRequest) error {
	if r.CapiProcessGuid == "" {
		return errors.New("CapiProcessGuid is required")
	}
	return nil
}
