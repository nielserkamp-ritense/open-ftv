package cerbos

// func TestNewController(t *testing.T) {
// 	t.Run("new controller", func(t *testing.T) {
// 		ctx, cancel := context.WithCancel(context.Background())
// 		defer cancel()
//
// 		h := slog2.NewDummyHandler(slog.LevelInfo)
// 		logger := slog.New(h)
//
// 		s := newService(t, "../../../../testdata/unittest/cerbos/conf/service_no_tls.yaml")
// 		require.NotNil(t, s)
//
// 		c := NewController(
// 			Config{
// 				Addr1: "passthrough:///" + s.GRPCAddr(),
// 				Addr2: "passthrough:///" + s.GRPCAddr(),
// 				User:  adminUser,
// 				Pswd:  adminPswd,
// 			},
// 			pdp.WithLogger(logger),
// 			pdp.WithStore("../../../../testdata/unittest/cerbos/policies", true),
// 			pdp.WithContext(ctx),
// 		)
// 		require.NotNil(t, c)
//
// 		assert.GreaterOrEqual(t, h.Count(), 4)
//
// 		c2, ok := c.(*controller)
// 		require.True(t, ok)
// 		require.NotNil(t, c2)
//
// 		assert.Equal(t, logger, c2.Logger())
// 		assert.NotNil(t, c2.engine)
// 		assert.NotNil(t, c2.admin)
// 		assert.NotNil(t, c2.info)
// 		assert.NotNil(t, c2.PAP())
// 		assert.NotNil(t, c2.policyIDs)
// 		assert.GreaterOrEqual(t, len(c2.policyIDs), 2)
// 	})
// }
//
// func TestNewControllerNoSvc(t *testing.T) {
// 	t.Run("new controller without svc", func(t *testing.T) {
// 		ctx, cancel := context.WithCancel(context.Background())
// 		defer cancel()
//
// 		h := slog2.NewDummyHandler(slog.LevelInfo)
// 		logger := slog.New(h)
//
// 		c := NewController(
// 			Config{
// 				Addr1: "https:///10.123.45.67:3593",
// 				Addr2: "https:///10.123.45.67:3593",
// 				User:  adminUser,
// 				Pswd:  adminPswd,
// 			},
// 			pdp.WithLogger(logger),
// 			pdp.WithStore("../../../../testdata/unittest/cerbos/policies", true),
// 			pdp.WithContext(ctx),
// 		)
// 		require.NotNil(t, c)
//
// 		assert.GreaterOrEqual(t, h.Count(), 5)
//
// 		c2, ok := c.(*controller)
// 		require.True(t, ok)
// 		require.NotNil(t, c2)
//
// 		assert.Equal(t, logger, c2.Logger())
// 		assert.NotNil(t, c2.engine)
// 		assert.NotNil(t, c2.admin)
// 		assert.Nil(t, c2.info)
// 		assert.NotNil(t, c2.PAP())
// 		assert.NotNil(t, c2.policyIDs)
// 		assert.Empty(t, c2.policyIDs)
// 	})
// }
//
// func TestNewControllerNoAdmin(t *testing.T) {
// 	t.Run("new controller without admin", func(t *testing.T) {
// 		ctx, cancel := context.WithCancel(context.Background())
// 		defer cancel()
//
// 		h := slog2.NewDummyHandler(slog.LevelInfo)
// 		logger := slog.New(h)
//
// 		s := newService(t, "../../../../testdata/unittest/cerbos/conf/service_no_tls.yaml")
// 		require.NotNil(t, s)
//
// 		c := NewController(
// 			Config{
// 				Addr1: "passthrough:///" + s.GRPCAddr(),
// 				User:  adminUser,
// 				Pswd:  adminPswd,
// 			},
// 			pdp.WithLogger(logger),
// 			pdp.WithStore("../../../../testdata/unittest/cerbos/policies", true),
// 			pdp.WithContext(ctx),
// 		)
// 		require.NotNil(t, c)
//
// 		assert.GreaterOrEqual(t, h.Count(), 5)
//
// 		c2, ok := c.(*controller)
// 		require.True(t, ok)
// 		require.NotNil(t, c2)
//
// 		assert.Equal(t, logger, c2.Logger())
// 		assert.NotNil(t, c2.engine)
// 		assert.Nil(t, c2.admin)
// 		assert.Nil(t, c2.info)
// 		assert.NotNil(t, c2.PAP())
// 		assert.NotNil(t, c2.policyIDs)
// 		assert.Empty(t, c2.policyIDs)
// 	})
// }
