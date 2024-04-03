package deviceapi

import (
	"context"
	"fmt"
	"time"

	pb "github.com/DIMO-Network/devices-api/pkg/grpc"
	gocache "github.com/patrickmn/go-cache"
	"google.golang.org/grpc"
)

const deviceTokenCacheKey = "udtoken_%d"

type Service struct {
	devicesConn *grpc.ClientConn
	memoryCache *gocache.Cache
}

// NewService API wrapper to call device-telemetry-api to get the userDevices associated with a userId over grpc
func NewService(devicesConn *grpc.ClientConn) *Service {
	c := gocache.New(8*time.Hour, 15*time.Minute)
	return &Service{devicesConn: devicesConn, memoryCache: c}
}

func (s *Service) GetUserDeviceByTokenID(ctx context.Context, tokenID int64) (*pb.UserDevice, error) {
	deviceClient := pb.NewUserDeviceServiceClient(s.devicesConn)
	var err error
	var userDevice *pb.UserDevice

	get, found := s.memoryCache.Get(fmt.Sprintf(deviceTokenCacheKey, tokenID))
	if found {
		userDevice = get.(*pb.UserDevice)
	} else {
		userDevice, err = deviceClient.GetUserDeviceByTokenId(ctx, &pb.GetUserDeviceByTokenIdRequest{
			TokenId: tokenID,
		})
		if err != nil {
			return nil, err
		}
		s.memoryCache.Set(fmt.Sprintf(deviceTokenCacheKey, tokenID), userDevice, time.Hour*24)
	}

	return userDevice, nil
}
