package certs

import (
	"crypto/x509"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/pkg/errors"
)

type CertManager struct {
	c configuration.Provider

	// isWindows is set to true if the application is running in a Windows
	// environment. This information is used to decide whether the stystem
	// certificate pool should be loaded or not as Windows does not support
	// this feature and will return an error.
	isWindows bool

	// cachePool is used to optimize performance and to avoid reloading the cert
	// pool at every request. The cache expires after its time-to-live has ended
	// or the an external action has invalidated the cache itself.
	cachePool *x509.CertPool

	// cacheTime contains the last update of the cache and is used to check
	// whether it has reached the end of its time-to-live.
	cacheTime time.Time

	// cacheCerts
	cacheCerts sync.Map

	requests int64
}

type CertCache struct {
	path string
	stat *os.FileInfo
}

func NewCertManager(c configuration.Provider) *CertManager {
	return &CertManager{
		c:         c,
		isWindows: runtime.GOOS == "windows",
	}
}

func (cm *CertManager) systemCertPool() (*x509.CertPool, error) {
	if cm.isWindows {
		return x509.NewCertPool(), nil
	}

	return x509.SystemCertPool()
}

func (cm *CertManager) additionalCerts() ([][]byte, error) {
	var certs [][]byte
	for _, cert := range cm.c.ProxyServeTransportCerts() {
		data, err := ioutil.ReadFile(cert)
		if err != nil {
			return nil, err
		}

		stat, err := os.Stat(cert)
		if err != nil {
			return nil, err
		}

		cm.cacheCerts.Store(cert, &CertCache{
			path: cert,
			stat: &stat,
		})

		certs = append(certs, data)
	}

	return certs, nil
}

func (cm *CertManager) cacheIsExpired() bool {
	if cm.c.ProxyServeTransportCacheTimeToLive() == 0 {
		return false
	}

	// If the last time the transport was accessed is beyond the cache TTL then
	// a refresh of the cache is forced whether the file was updated or not.
	// This ensures that the cache is always refreshed at fixed intervals
	// regardless of the environment. If the TTL is set to 0 skip the check.
	return time.Since(cm.cacheTime) > cm.c.ProxyServeTransportCacheTimeToLive()
}

func (cm *CertManager) lookupCertificate(cert string) (*CertCache, bool) {
	cacheItem, ok := cm.cacheCerts.Load(cert)
	if !ok {
		return nil, false
	}

	cache, ok := cacheItem.(*CertCache)
	return cache, ok
}

func (cm *CertManager) isCertificateValid(cert string) (bool, error) {
	cache, ok := cm.lookupCertificate(cert)
	if !ok {
		return false, nil
	}

	refreshFrequency := cm.c.ProxyServeTransportCacheRefreshFrequency()
	if refreshFrequency <= 0 {
		return true, nil
	}

	if cm.requests%int64(refreshFrequency) != 0 {
		return true, nil
	}

	stat, err := os.Stat(cert)
	if err != nil {
		return false, err
	}

	if stat.Size() != (*cache.stat).Size() || stat.ModTime() != (*cache.stat).ModTime() {
		return true, nil
	}

	return false, nil
}

func (cm *CertManager) cacheIsValid() (bool, error) {
	for _, cert := range cm.c.ProxyServeTransportCerts() {
		valid, err := cm.isCertificateValid(cert)
		if err != nil {
			return false, err
		}

		if !valid {
			return false, nil
		}
	}

	return true, nil
}

func (cm *CertManager) CertPool() (*x509.CertPool, error) {
	atomic.AddInt64(&cm.requests, 1)

	cacheValid, err := cm.cacheIsValid()
	if err != nil {
		return nil, err
	}

	if !cm.cacheIsExpired() && cacheValid {
		return cm.cachePool, nil
	}

	certs, err := cm.additionalCerts()
	if err != nil {
		return nil, err
	}

	pool, err := cm.systemCertPool()
	if err != nil {
		return nil, err
	}

	for _, data := range certs {
		if ok := pool.AppendCertsFromPEM(data); !ok {
			return nil, errors.New("No certs appended, only system certs present, did you specify the correct cert file?")
		}
	}

	return pool, nil
}
