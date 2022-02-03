package object

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/nspcc-dev/neofs-api-go/v2/object"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
	"strconv"
	"time"
)

//lifted from the HTTP Gateway code

//https://github.com/nspcc-dev/neofs-http-gw/blob/master/uploader/filter.go#L72

const (
	UserAttributeHeaderPrefix = "X-Attribute-"
	SystemAttributePrefix     = "__NEOFS__"

	ExpirationDurationAttr  = SystemAttributePrefix + "EXPIRATION_DURATION"
	ExpirationTimestampAttr = SystemAttributePrefix + "EXPIRATION_TIMESTAMP"
	ExpirationRFC3339Attr   = SystemAttributePrefix + "EXPIRATION_RFC3339"
)
type epochDurations struct {
	currentEpoch  uint64
	msPerBlock    int64
	blockPerEpoch uint64
}

func getEpochDurations(ctx context.Context, p pool.Pool) (*epochDurations, error) {
	if conn, _, err := p.Connection(); err != nil {
		return nil, err
	} else if networkInfoRes, err := conn.NetworkInfo(ctx); err != nil {
		return nil, err
	} else if err = apistatus.ErrFromStatus(networkInfoRes.Status()); err != nil {
		return nil, err
	} else {
		networkInfo := networkInfoRes.Info()
		res := &epochDurations{
			currentEpoch: networkInfo.CurrentEpoch(),
			msPerBlock:   networkInfo.MsPerBlock(),
		}

		networkInfo.NetworkConfig().IterateParameters(func(parameter *netmap.NetworkParameter) bool {
			if string(parameter.Key()) == "EpochDuration" {
				data := make([]byte, 8)
				copy(data, parameter.Value())
				res.blockPerEpoch = binary.LittleEndian.Uint64(data)
				return true
			}
			return false
		})
		if res.blockPerEpoch == 0 {
			return nil, fmt.Errorf("not found param: EpochDuration")
		}
		return res, nil
	}
}
//this runs through the provided headers and converts a human readable time format to the correct header format for an expiration
func prepareExpirationHeader(headers map[string]string, epochDurations *epochDurations) error {
	expirationInEpoch := headers[object.SysAttributeExpEpoch]

	if timeRFC3339, ok := headers[ExpirationRFC3339Attr]; ok {
		expTime, err := time.Parse(time.RFC3339, timeRFC3339)
		if err != nil {
			return fmt.Errorf("couldn't parse value %s of header %s", timeRFC3339, ExpirationRFC3339Attr)
		}

		now := time.Now().UTC()
		if expTime.Before(now) {
			return fmt.Errorf("value %s of header %s must be in the future", timeRFC3339, ExpirationRFC3339Attr)
		}
		updateExpirationHeader(headers, epochDurations, expTime.Sub(now))
		delete(headers, ExpirationRFC3339Attr)
	}

	if timestamp, ok := headers[ExpirationTimestampAttr]; ok {
		value, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return fmt.Errorf("couldn't parse value %s of header %s", timestamp, ExpirationTimestampAttr)
		}
		expTime := time.Unix(value, 0)

		now := time.Now()
		if expTime.Before(now) {
			return fmt.Errorf("value %s of header %s must be in the future", timestamp, ExpirationTimestampAttr)
		}
		updateExpirationHeader(headers, epochDurations, expTime.Sub(now))
		delete(headers, ExpirationTimestampAttr)
	}

	if duration, ok := headers[ExpirationDurationAttr]; ok {
		expDuration, err := time.ParseDuration(duration)
		if err != nil {
			return fmt.Errorf("couldn't parse value %s of header %s", duration, ExpirationDurationAttr)
		}
		if expDuration <= 0 {
			return fmt.Errorf("value %s of header %s must be positive", expDuration, ExpirationDurationAttr)
		}
		updateExpirationHeader(headers, epochDurations, expDuration)
		delete(headers, ExpirationDurationAttr)
	}

	if expirationInEpoch != "" {
		headers[object.SysAttributeExpEpoch] = expirationInEpoch
	}

	return nil
}

func updateExpirationHeader(headers map[string]string, durations *epochDurations, expDuration time.Duration) {
	epochDuration := durations.msPerBlock * int64(durations.blockPerEpoch)
	numEpoch := expDuration.Milliseconds() / epochDuration
	headers[object.SysAttributeExpEpoch] = strconv.FormatInt(int64(durations.currentEpoch)+numEpoch, 10)
}
