package taskmanager

import "testing"

func testResQuotaConf() *Conf {
	conf := clone(CONF)
	conf.StorageKeyPrefix = "testtaskmanager"
	conf.EtcdIps = []string{"localhost:2379"}
	conf.ResourceQuotaName = "test-spark"
	return conf
}

func TestInit(t *testing.T) {
	conf := testResQuotaConf()
	m, err := NewResourceQuotaManager(conf)
	if err != nil {
		t.Fatalf("%v", err)
	}
	quota := ResourceQuota{LimitCPU: 1, LimitMem: 100}
	err = m.Init(conf.ResourceQuotaName, &quota)
	if err != nil {
		t.Fatalf("%v", err)
	}
}
