// +build unit

package txsentry

// 
// func TestPendingJobs_Execute(t *testing.T) {
// 	ctx := context.Background()
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()
// 
// 	apiClient := mock.NewMockOrchestrateClient(ctrl)
// 	sendRetryJobUseCase := mocks.NewMockSendRetryJob(ctrl)
// 	retrySessionState := storemocks.NewMockRetrySessions(ctrl)
// 	logger := log.NewLogger()
// 
// 	expectedErr := fmt.Errorf("expected_err")
// 
// 	t.Run("should start and exit gracefully after running retry session successfully", func(t *testing.T) {
// 		parentJob := testdata.FakeJob()
// 		childJob := testdata.FakeJob()
// 		parentJob.InternalData.RetryInterval = time.Second
// 
// 		apiClient.EXPECT().SearchJob(gomock.Any(), gomock.Any()).Return([]*api.JobResponse{}, nil)
// 		retrySessionState.EXPECT().Has(gomock.Any(), parentJob.UUID).Return(true)
// 		prevCall := sendRetryJobUseCase.EXPECT().Execute(gomock.Any(), parentJob, parentJob.UUID, 0).Return(childJob.UUID, nil)
// 		sendRetryJobUseCase.EXPECT().Execute(gomock.Any(), parentJob, childJob.UUID, 1).After(prevCall).
// 			Return("", nil)
// 
// 		usecase := NewRetryJobSession(apiClient, sendRetryJobUseCase, retrySessionState, parentJob, logger)
// 		go func() {
// 			err := usecase.Start(ctx)
// 			require.NoError(t, err)
// 		}()
// 
// 		time.Sleep(parentJob.InternalData.RetryInterval * 2)
// 		err := usecase.Stop()
// 		assert.NoError(t, err)
// 	})
// 
// 	t.Run("should start and exit gracefully after running retry session successfully", func(t *testing.T) {
// 		parentJob := testdata.FakeJob()
// 		childJob := testdata.FakeJob()
// 		parentJob.InternalData.RetryInterval = time.Second
// 
// 		apiClient.EXPECT().SearchJob(gomock.Any(), gomock.Any()).Return([]*api.JobResponse{}, nil)
// 		prevCall := sendRetryJobUseCase.EXPECT().Execute(gomock.Any(), parentJob, parentJob.UUID, 0).Return(childJob.UUID, nil)
// 		sendRetryJobUseCase.EXPECT().Execute(gomock.Any(), parentJob, childJob.UUID, 1).After(prevCall).
// 			Return("", nil)
// 		retrySessionState.EXPECT().Remove(gomock.Any(), parentJob.UUID)
// 
// 		usecase := NewRetryJobSession(apiClient, sendRetryJobUseCase, retrySessionState, parentJob, logger)
// 		err := usecase.Start(ctx)
// 		require.NoError(t, err)
// 		time.Sleep(parentJob.InternalData.RetryInterval * 2)
// 		err = usecase.Stop()
// 		assert.NoError(t, err)
// 	})
// 
// 	t.Run("should start and exit with errors after running retry session fails", func(t *testing.T) {
// 		parentJob := testdata.FakeJob()
// 		parentJob.InternalData.RetryInterval = time.Second
// 
// 		retrySessionState.EXPECT().Add(gomock.Any(), parentJob.UUID, parentJob)
// 		apiClient.EXPECT().SearchJob(gomock.Any(), gomock.Any()).Return([]*api.JobResponse{}, nil)
// 		sendRetryJobUseCase.EXPECT().Execute(gomock.Any(), parentJob, parentJob.UUID, 0).Return("", expectedErr)
// 		retrySessionState.EXPECT().Remove(gomock.Any(), parentJob.UUID)
// 
// 		usecase := NewRetryJobSession(apiClient, sendRetryJobUseCase, retrySessionState, parentJob, logger)
// 		err := usecase.Start(ctx)
// 		require.NoError(t, err)
// 		time.Sleep(parentJob.InternalData.RetryInterval * 2)
// 		err = usecase.Stop()
// 		assert.Error(t, err)
// 		assert.Equal(t, expectedErr, err)
// 	})
// }
