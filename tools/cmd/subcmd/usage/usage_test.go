package usage

import "testing"

func TestUsage(t *testing.T) {
	sumPods, sumNS, err := doRun("default", "/Users/xubei1/.kube/config")
	if err != nil {
		t.Fatalf("%v", err)
	}

	t.Logf("pods sum %v, ns sum %v", sumPods, sumNS)
}
