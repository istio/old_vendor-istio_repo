# frozen_string_literal: true

require 'copilot/protos/cloud_controller_pb'
require 'copilot/protos/cloud_controller_services_pb'

module Cloudfoundry
  module Copilot
    class Client

      attr_reader :host, :port

      def initialize(host:, port:, client_ca_file:, client_key_file:, client_chain_file:, timeout: 5)
        @host = host
        @port = port
        @url = "#{host}:#{port}"
        @timeout = timeout
        @client_ca = File.open(client_ca_file).read
        @client_key = File.open(client_key_file).read
        @client_chain = File.open(client_chain_file).read
      end

      def health
        request = Api::HealthRequest.new
        service.health(request)
      end

      def upsert_route(guid:, host:)
        route = Api::Route.new(guid: guid, host: host)
        request = Api::UpsertRouteRequest.new(route: route)
        service.upsert_route(request)
      end

      def delete_route(guid:)
        request = Api::DeleteRouteRequest.new(guid: guid)
        service.delete_route(request)
      end

      def map_route(capi_process_guid:, route_guid:)
        route_mapping = Api::RouteMapping.new(capi_process_guid: capi_process_guid, route_guid: route_guid)
        request = Api::MapRouteRequest.new(route_mapping: route_mapping)
        service.map_route(request)
      end

      def unmap_route(capi_process_guid:, route_guid:)
        route_mapping = Api::RouteMapping.new(capi_process_guid: capi_process_guid, route_guid: route_guid)
        request = Api::UnmapRouteRequest.new(route_mapping: route_mapping)
        service.unmap_route(request)
      end

      def upsert_capi_diego_process_association(capi_process_guid:, diego_process_guids:)
        request = Api::UpsertCapiDiegoProcessAssociationRequest.new(
          capi_diego_process_association: {
            capi_process_guid: capi_process_guid,
            diego_process_guids: diego_process_guids
          })

        service.upsert_capi_diego_process_association(request)
      end

      def delete_capi_diego_process_association(capi_process_guid:)
        request = Api::DeleteCapiDiegoProcessAssociationRequest.new(
          capi_process_guid: capi_process_guid
        )
        service.delete_capi_diego_process_association(request)
      end

      # untested - this will change a lot and no one is using it yet
      def bulk_sync(routes:, route_mappings:)
        routes.map! {|route| Api::UpsertRouteRequest.new(route: route)}
        route_mappings.map! {|mapping| Api::MapRouteRequest.new(route_mapping: mapping)}

        request = Api::BulkSyncRequest.new(routes: routes, route_mappings: route_mappings)
        service.bulk_sync(request)
      end

      private

      def tls_credentials
        @tls_credentials ||= GRPC::Core::ChannelCredentials.new(@client_ca, @client_key, @client_chain)
      end

      def service
        @service ||= Api::CloudControllerCopilot::Stub.new(@url, tls_credentials, timeout: @timeout)
      end
    end
  end
end
