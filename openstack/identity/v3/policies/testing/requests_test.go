package testing

import (
	"fmt"
	"testing"

	"github.com/gophercloud/gophercloud/openstack/identity/v3/policies"
	"github.com/gophercloud/gophercloud/pagination"
	th "github.com/gophercloud/gophercloud/testhelper"
	"github.com/gophercloud/gophercloud/testhelper/client"
)

func TestListPolicies(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleListPoliciesSuccessfully(t)

	count := 0
	err := policies.List(client.ServiceClient(), nil).EachPage(func(page pagination.Page) (bool, error) {
		count++

		actual, err := policies.ExtractPolicies(page)
		th.AssertNoErr(t, err)

		th.CheckDeepEquals(t, ExpectedPoliciesSlice, actual)

		return true, nil
	})
	th.AssertNoErr(t, err)
	th.CheckEquals(t, count, 1)
}

func TestListPoliciesAllPages(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleListPoliciesSuccessfully(t)

	allPages, err := policies.List(client.ServiceClient(), nil).AllPages()
	th.AssertNoErr(t, err)
	actual, err := policies.ExtractPolicies(allPages)
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, ExpectedPoliciesSlice, actual)
}

func TestListPoliciesWithFilter(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleListPoliciesSuccessfully(t)

	listOpts := policies.ListOpts{
		Type: "application/json",
	}
	allPages, err := policies.List(client.ServiceClient(), listOpts).AllPages()
	th.AssertNoErr(t, err)
	actual, err := policies.ExtractPolicies(allPages)
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, []policies.Policy{SecondPolicy}, actual)
}

func TestCreatePolicy(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleCreatePolicySuccessfully(t)

	createOpts := policies.CreateOpts{
		Type: "application/json",
		Blob: []byte("{'bar_user': 'role:network-user'}"),
		Extra: map[string]interface{}{
			"description": "policy for bar_user",
		},
	}

	actual, err := policies.Create(client.ServiceClient(), createOpts).Extract()
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, SecondPolicy, *actual)
}

func TestCreatePolicyTypeLengthCheck(t *testing.T) {
	// strGenerator generates a string of fixed length filled with '0'
	strGenerator := func(length int) string {
		return fmt.Sprintf(fmt.Sprintf("%%0%dd", length), 0)
	}

	type test struct {
		length  int
		wantErr bool
	}

	tests := []test{
		{100, false},
		{255, false},
		{256, true},
		{300, true},
	}

	createOpts := policies.CreateOpts{
		Blob: []byte("{'bar_user': 'role:network-user'}"),
	}

	for _, _test := range tests {
		createOpts.Type = strGenerator(_test.length)
		if len(createOpts.Type) != _test.length {
			t.Fatal("function strGenerator does not work properly")
		}

		_, err := createOpts.ToPolicyCreateMap()
		if !_test.wantErr {
			th.AssertNoErr(t, err)
		} else {
			switch _t := err.(type) {
			case nil:
				t.Fatal("error expected but got a nil")
			case policies.StringFieldLengthExceedsLimit:
			default:
				t.Fatalf("unexpected error type: [%T]", _t)
			}
		}
	}
}