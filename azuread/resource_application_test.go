package azuread

import (
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-azuread/azuread/helpers/ar"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAzureADApplication_basic(t *testing.T) {
	resourceName := "azuread_application.test"
	id := uuid.New().String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckADApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccADApplication_basic(id),
				Check: resource.ComposeTestCheckFunc(
					testCheckADApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("acctest%s", id)),
					resource.TestCheckResourceAttr(resourceName, "homepage", fmt.Sprintf("https://acctest%s", id)),
					resource.TestCheckResourceAttrSet(resourceName, "application_id"),
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

func TestAccAzureADApplication_availableToOtherTenants(t *testing.T) {
	resourceName := "azuread_application.test"
	id := uuid.New().String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckADApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccADApplication_availableToOtherTenants(id),
				Check: resource.ComposeTestCheckFunc(
					testCheckADApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "available_to_other_tenants", "true"),
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

func TestAccAzureADApplication_complete(t *testing.T) {
	resourceName := "azuread_application.test"
	id := uuid.New().String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckADApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccADApplication_complete(id),
				Check: resource.ComposeTestCheckFunc(
					testCheckADApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("acctest%s", id)),
					resource.TestCheckResourceAttr(resourceName, "homepage", fmt.Sprintf("https://homepage-%s", id)),
					resource.TestCheckResourceAttr(resourceName, "identifier_uris.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reply_urls.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_id"),
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

func TestAccAzureADApplication_update(t *testing.T) {
	resourceName := "azuread_application.test"
	id := uuid.New().String()
	updatedId := uuid.New().String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckADApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccADApplication_basic(id),
				Check: resource.ComposeTestCheckFunc(
					testCheckADApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("acctest%s", id)),
					resource.TestCheckResourceAttr(resourceName, "homepage", fmt.Sprintf("https://acctest%s", id)),
					resource.TestCheckResourceAttr(resourceName, "identifier_uris.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reply_urls.#", "0"),
				),
			},
			{
				Config: testAccADApplication_complete(updatedId),
				Check: resource.ComposeTestCheckFunc(
					testCheckADApplicationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("acctest%s", updatedId)),
					resource.TestCheckResourceAttr(resourceName, "homepage", fmt.Sprintf("https://homepage-%s", updatedId)),
					resource.TestCheckResourceAttr(resourceName, "identifier_uris.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reply_urls.#", "1"),
				),
			},
		},
	})
}

func testCheckADApplicationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %q", name)
		}

		client := testAccProvider.Meta().(*ArmClient).applicationsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		resp, err := client.Get(ctx, rs.Primary.ID)

		if err != nil {
			if ar.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Azure AD Application %q does not exist", rs.Primary.ID)
			}
			return fmt.Errorf("Bad: Get on Azure AD applicationsClient: %+v", err)
		}

		return nil
	}
}

func testCheckADApplicationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azuread_application" {
			continue
		}

		client := testAccProvider.Meta().(*ArmClient).applicationsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext
		resp, err := client.Get(ctx, rs.Primary.ID)

		if err != nil {
			if ar.ResponseWasNotFound(resp.Response) {
				return nil
			}

			return err
		}

		return fmt.Errorf("Azure AD Application still exists:\n%#v", resp)
	}

	return nil
}

func testAccADApplication_basic(id string) string {
	return fmt.Sprintf(`
resource "azuread_application" "test" {
  name = "acctest%s"
}
`, id)
}

func testAccADApplication_availableToOtherTenants(id string) string {
	return fmt.Sprintf(`
resource "azuread_application" "test" {
  name                       = "acctest%s"
  identifier_uris            = ["https://%s.hashicorptest.com"]
  available_to_other_tenants = true
}
`, id, id)
}

func testAccADApplication_complete(id string) string {
	return fmt.Sprintf(`
resource "azuread_application" "test" {
  name                       = "acctest%s"
  homepage                   = "https://homepage-%s"
  identifier_uris            = ["http://%s.hashicorptest.com"]
  reply_urls                 = ["http://%s.hashicorptest.com"]
  oauth2_allow_implicit_flow = true
}
`, id, id, id, id)
}
