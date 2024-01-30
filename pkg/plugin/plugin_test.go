package plugin

//func TestSetWeight(t *testing.T) {
//	assert := assert.New(t)
//
//	testCases := []struct {
//		rollout                *v1alpha1.Rollout
//		desiredWeight          int32
//		additionalDestinations []v1alpha1.WeightDestination
//		expectedError          pluginTypes.RpcError
//	}{
//		// Add your test cases here
//	}
//
//	for _, testCase := range testCases {
//		p := &RpcPlugin{
//			SplitterClient: nil,
//			ResolverClient: nil,
//			IsTest:         true,
//		}
//		err := p.SetWeight(testCase.rollout, testCase.desiredWeight, testCase.additionalDestinations)
//		assert.Equal(testCase.expectedError, err, "errors should be equal")
//	}
//}
