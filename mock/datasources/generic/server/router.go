package server

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"

	handle1 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/types"
	handle2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/handlers/fiber"
)

func (s *service) initRoutes(_ context.Context, svc *fiber.App) {
	svc.Get("/healthz", handle1.HealthZ)

	v1 := svc.Group("/v1")
	meta := v1.Group("/meta")
	data := v1.Group("/data")

	spaceHandler := handle2.NewDataspaceHandler(s.db, s.logger)
	meta.Get(handle2.PathDataspace, spaceHandler.GetDataspace)
	meta.Put(handle2.PathDataspace, spaceHandler.PutDataspace)
	meta.Post(handle2.PathDataspace, spaceHandler.PostDataspace)
	meta.Delete(handle2.PathDataspace, spaceHandler.DeleteDataspace)

	sourceHandler := handle2.NewDatasourceHandler(s.db, s.logger)
	meta.Get(handle2.PathDatasources, sourceHandler.GetDatasources)
	meta.Get(handle2.PathDatasource, sourceHandler.GetDatasource)
	meta.Put(handle2.PathDatasource, sourceHandler.PutDatasource)
	meta.Post(handle2.PathDatasource, sourceHandler.PostDatasource)
	meta.Delete(handle2.PathDatasource, sourceHandler.DeleteDatasource)

	tableHandler := handle2.NewTableHandler(s.db, s.logger)
	meta.Get(handle2.PathTables, tableHandler.GetTables)
	meta.Get(handle2.PathTable, tableHandler.GetTable)
	meta.Put(handle2.PathTable, tableHandler.PutTable)
	meta.Post(handle2.PathTable, tableHandler.PostTable)
	meta.Delete(handle2.PathTable, tableHandler.DeleteTable)

	dataHandler := handle2.NewTableDataHandler(s.db, s.logger)
	data.Get(handle2.PathTableRecords, dataHandler.GetRecords)
	data.Get(handle2.PathTableRecord, dataHandler.GetRecord)
	data.Put(handle2.PathTableRecord, dataHandler.PutRecord)
	data.Post(handle2.PathTableRecord, dataHandler.PostRecord)
	data.Delete(handle2.PathTableRecord, dataHandler.DeleteRecord)

	versions := map[uint8]fiber.Router{1: v1}

	s.db.IterateEndpoints(func(def *schema.Endpoint) {
		version := versions[def.Version]
		if version == nil {
			version = svc.Group(fmt.Sprintf("/v%d", def.Version))
			versions[def.Version] = version
		}

		h := handle2.NewEndpointHandler(s.db, s.logger, def)

		switch def.CalledAs {
		case types.GetMethod:
			version.Get(def.Path, h.Handle)
		case types.PostMethod:
			version.Post(def.Path, h.Handle)
		case types.PutMethod:
			version.Put(def.Path, h.Handle)
		case types.DeleteMethod:
			version.Delete(def.Path, h.Handle)
		}
	})
}
