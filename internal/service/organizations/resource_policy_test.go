package organizations_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var policy organizations.ResourcePolicy
	resourceName := "aws_organizations_resource_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, resourceName, &policy),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "organizations", regexp.MustCompile("resourcepolicy/o-.+/rp-.+$")),
					resource.TestCheckResourceAttrSet(resourceName, "content"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckResourcePolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_organizations_resource_policy" {
				continue
			}

			_, err := tforganizations.FindResourcePolicy(ctx, conn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Organizations Resource Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckResourcePolicyExists(ctx context.Context, n string, v *organizations.ResourcePolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsConn(ctx)

		output, err := tforganizations.FindResourcePolicy(ctx, conn)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccResourcePolicyConfig_basic() string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), `
data "aws_caller_identity" "delegated" {
  provider = "awsalternate"
}

resource "aws_organizations_resource_policy" "test" {
  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "DelegatingNecessaryDescribeListActions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "${data.aws_caller_identity.delegated.arn}"
      },
      "Action": [
        "organizations:DescribeOrganization",
        "organizations:DescribeOrganizationalUnit",
        "organizations:DescribeAccount",
        "organizations:DescribePolicy",
        "organizations:DescribeEffectivePolicy",
        "organizations:ListRoots",
        "organizations:ListOrganizationalUnitsForParent",
        "organizations:ListParents",
        "organizations:ListChildren",
        "organizations:ListAccounts",
        "organizations:ListAccountsForParent",
        "organizations:ListPolicies",
        "organizations:ListPoliciesForTarget",
        "organizations:ListTargetsForPolicy",
        "organizations:ListTagsForResource"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}
`)
}
