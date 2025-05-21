package server

import (
	"context"

	"github.com/gofiber/fiber/v2"

	handle1 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers/fiber"
	handle2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/handlers/fiber"
)

func (s *service) initRoutes(_ context.Context, svc *fiber.App) {
	svc.Get("/healthz", handle1.HealthZ)

	v1 := svc.Group("/v1")

	spaces := handle2.NewDataspaceHandler(s.db, s.logger)
	v1.Get(handle2.PathDataspace, spaces.GetDataspace)
	v1.Put(handle2.PathDataspace, spaces.PutDataspace)
	v1.Post(handle2.PathDataspace, spaces.PostDataspace)
	v1.Delete(handle2.PathDataspace, spaces.DeleteDataspace)

	sources := handle2.NewDatasourceHandler(s.db, s.logger)
	v1.Get(handle2.PathDatasources, sources.GetDatasources)
	v1.Get(handle2.PathDatasource, sources.GetDatasource)
	v1.Put(handle2.PathDatasource, sources.PutDatasource)
	v1.Post(handle2.PathDatasource, sources.PostDatasource)
	v1.Delete(handle2.PathDatasource, sources.DeleteDatasource)

	tables := handle2.NewTableHandler(s.db, s.logger)
	v1.Get(handle2.PathTables, tables.GetTables)
	v1.Get(handle2.PathTable, tables.GetTable)
	v1.Put(handle2.PathTable, tables.PutTable)
	v1.Post(handle2.PathTable, tables.PostTable)
	v1.Delete(handle2.PathTable, tables.DeleteTable)

	data := handle2.NewTableDataHandler(s.db, s.logger)
	v1.Get(handle2.PathTableRecords, data.GetRecords)
	v1.Get(handle2.PathTableRecord, data.GetRecord)
	v1.Put(handle2.PathTableRecord, data.PutRecord)
	v1.Post(handle2.PathTableRecord, data.PostRecord)
	v1.Delete(handle2.PathTableRecord, data.DeleteRecord)
}
