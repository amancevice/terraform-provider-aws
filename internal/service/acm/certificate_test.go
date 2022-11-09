package acm_test

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfacm "github.com/hashicorp/terraform-provider-aws/internal/service/acm"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccACMCertificate_emailValidation(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var v acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(domain, acm.ValidationMethodEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "not_after", ""),
					resource.TestCheckResourceAttr(resourceName, "not_before", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domain),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypeAmazonIssued),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityIneligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "validation_emails.#", "0"),
					resource.TestMatchResourceAttr(resourceName, "validation_emails.0", regexp.MustCompile(`^[^@]+@.+$`)),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodEmail),
					resource.TestCheckResourceAttr(resourceName, "validation_option.#", "0"),
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

func TestAccACMCertificate_dnsValidation(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var v acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(domain, acm.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "not_after", ""),
					resource.TestCheckResourceAttr(resourceName, "not_before", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domain),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypeAmazonIssued),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityIneligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
					resource.TestCheckResourceAttr(resourceName, "validation_option.#", "0"),
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

func TestAccACMCertificate_root(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	var v acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(rootDomain, acm.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
					resource.TestCheckResourceAttr(resourceName, "validation_option.#", "0"),
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

func TestAccACMCertificate_validationOptions(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var v acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_validationOptions(rootDomain, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domain),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "validation_emails.#", "0"),
					resource.TestMatchResourceAttr(resourceName, "validation_emails.0", regexp.MustCompile(`^[^@]+@.+$`)),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodEmail),
					resource.TestCheckResourceAttr(resourceName, "validation_option.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"validation_option"},
			},
		},
	})
}

func TestAccACMCertificate_privateCertificate_renewable(t *testing.T) {
	certificateAuthorityResourceName := "aws_acmpca_certificate_authority.test"
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	var v1, v2, v3, v4 acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_authority_arn", certificateAuthorityResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_after"),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_before"),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityIneligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "validation_option.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMConn

					_, err := conn.ExportCertificate(&acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					})
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.StringValue(v1.CertificateArn), err)
					}
				},
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v2),
					testAccCheckCertificateNotRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityEligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMConn

					ctx := context.Background()
					_, err := conn.RenewCertificateWithContext(ctx, &acm.RenewCertificateInput{
						CertificateArn: v1.CertificateArn,
					})
					if err != nil {
						t.Fatalf("renewing ACM Certificate (%s): %s", aws.StringValue(v1.CertificateArn), err)
					}
				},
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityEligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status", acm.RenewalStatusPendingAutoRenewal),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status_reason", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "renewal_summary.0.updated_at"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMConn

					ctx := context.Background()
					_, err := tfacm.WaitCertificateRenewed(ctx, conn, aws.StringValue(v1.CertificateArn), tfacm.CertificateRenewalTimeout)
					if err != nil {
						t.Fatalf("waiting for ACM Certificate (%s) renewal: %s", aws.StringValue(v1.CertificateArn), err)
					}
				},
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v4),
					testAccCheckCertificateRenewed(&v3, &v4),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityIneligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status", acm.RenewalStatusSuccess),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status_reason", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "renewal_summary.0.updated_at"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
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

func TestAccACMCertificate_privateCertificate_noRenewalPermission(t *testing.T) {
	certificateAuthorityResourceName := "aws_acmpca_certificate_authority.test"
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	var v1, v2, v3 acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_noRenewalPermission(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_authority_arn", certificateAuthorityResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_after"),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_before"),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityIneligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "validation_option.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMConn

					_, err := conn.ExportCertificate(&acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					})
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.StringValue(v1.CertificateArn), err)
					}
				},
				Config: testAccCertificateConfig_privateCertificate_noRenewalPermission(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityEligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMConn

					ctx := context.Background()
					_, err := conn.RenewCertificateWithContext(ctx, &acm.RenewCertificateInput{
						CertificateArn: v1.CertificateArn,
					})
					if err != nil {
						t.Fatalf("renewing ACM Certificate (%s): %s", aws.StringValue(v1.CertificateArn), err)
					}
				},
				Config: testAccCertificateConfig_privateCertificate_noRenewalPermission(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityEligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status", acm.RenewalStatusPendingAutoRenewal),
					acctest.CheckResourceAttrRFC3339(resourceName, "renewal_summary.0.updated_at"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
			{
				PreConfig: func() {
					time.Sleep(tfacm.CertificateRenewalTimeout)
				},
				Config: testAccCertificateConfig_privateCertificate_noRenewalPermission(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityEligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status", acm.RenewalStatusPendingAutoRenewal),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status_reason", acm.FailureReasonPcaAccessDenied),
					acctest.CheckResourceAttrRFC3339(resourceName, "renewal_summary.0.updated_at"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
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

func TestAccACMCertificate_privateCertificate_pendingRenewalGoDuration(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	duration := (395 * 24 * time.Hour).String()
	var v1, v2 acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityIneligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMConn

					_, err := conn.ExportCertificate(&acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					})
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.StringValue(v1.CertificateArn), err)
					}
				},
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is true and `renewal_eligibility` is `ELIGIBLE` after exporting
				// before actually performing the renewal.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v2),
					testAccCheckCertificateRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityIneligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status", acm.RenewalStatusSuccess),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status_reason", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "renewal_summary.0.updated_at"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
		},
	})
}

func TestAccACMCertificate_privateCertificate_pendingRenewalRFC3339Duration(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	duration := "P395D"
	var v1, v2 acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityIneligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMConn

					_, err := conn.ExportCertificate(&acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					})
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.StringValue(v1.CertificateArn), err)
					}
				},
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is true and `renewal_eligibility` is `ELIGIBLE` after exporting
				// before actually performing the renewal.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v2),
					testAccCheckCertificateRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityIneligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status", acm.RenewalStatusSuccess),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status_reason", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "renewal_summary.0.updated_at"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
		},
	})
}

func TestAccACMCertificate_privateCertificate_addEarlyRenewalPast(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	duration := (395 * 24 * time.Hour).String()
	var v1, v2, v3 acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityIneligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMConn

					_, err := conn.ExportCertificate(&acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					})
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.StringValue(v1.CertificateArn), err)
					}
				},
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is false and `renewal_eligibility` is `ELIGIBLE` after exporting.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v2),
					testAccCheckCertificateNotRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityEligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
			{
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is true after setting `early_renewal_duration`.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v3),
					testAccCheckCertificateRenewed(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityIneligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status", acm.RenewalStatusSuccess),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status_reason", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "renewal_summary.0.updated_at"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
		},
	})
}

func TestAccACMCertificate_privateCertificate_addEarlyRenewalPastIneligible(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	duration := (395 * 24 * time.Hour).String()
	var v1, v2 acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityIneligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
			{
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is true after setting `early_renewal_duration`.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v2),
					testAccCheckCertificateNotRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityIneligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
		},
	})
}

func TestAccACMCertificate_privateCertificate_addEarlyRenewalFuture(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	duration := (90 * 24 * time.Hour).String()
	var v1, v2, v3 acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityIneligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMConn

					_, err := conn.ExportCertificate(&acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					})
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.StringValue(v1.CertificateArn), err)
					}
				},
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is false and `renewal_eligibility` is `ELIGIBLE` after exporting.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v2),
					testAccCheckCertificateNotRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityEligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
			{
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is true after setting `early_renewal_duration`.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v3),
					testAccCheckCertificateNotRenewed(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityEligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
		},
	})
}

func TestAccACMCertificate_privateCertificate_updateEarlyRenewalFuture(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	duration := (395 * 24 * time.Hour).String()
	durationUpdated := (90 * 24 * time.Hour).String()
	var v1, v2 acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityIneligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMConn

					_, err := conn.ExportCertificate(&acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					})
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.StringValue(v1.CertificateArn), err)
					}
				},
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is false and `renewal_eligibility` is `ELIGIBLE` after exporting.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, durationUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v2),
					testAccCheckCertificateNotRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", durationUpdated),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityEligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
		},
	})
}

func TestAccACMCertificate_privateCertificate_removeEarlyRenewal(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	duration := (395 * 24 * time.Hour).String()
	var v1, v2 acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityIneligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMConn

					_, err := conn.ExportCertificate(&acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					})
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.StringValue(v1.CertificateArn), err)
					}
				},
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is false and `renewal_eligibility` is `ELIGIBLE` after exporting.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v2),
					testAccCheckCertificateNotRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityEligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypePrivate),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
		},
	})
}

// TestAccACMCertificate_Root_trailingPeriod updated in 3.0 to account for domain_name plan-time validation
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13510
func TestAccACMCertificate_Root_trailingPeriod(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := fmt.Sprintf("%s.", rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateConfig_basic(domain, acm.ValidationMethodDns),
				ExpectError: regexp.MustCompile(`invalid value for domain_name \(cannot end with a period\)`),
			},
		},
	})
}

func TestAccACMCertificate_rootAndWildcardSan(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)
	var v acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(rootDomain, strconv.Quote(wildcardDomain), acm.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          wildcardDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
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

func TestAccACMCertificate_SubjectAlternativeNames_emptyString(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(""), acm.ValidationMethodDns),
				ExpectError: regexp.MustCompile(`expected length`),
			},
		},
	})
}

func TestAccACMCertificate_San_single(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var v acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(sanDomain), acm.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          sanDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domain),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", sanDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
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

func TestAccACMCertificate_San_multiple(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain1 := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain2 := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var v acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(domain, fmt.Sprintf("%q, %q", sanDomain1, sanDomain2), acm.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          sanDomain1,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          sanDomain2,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domain),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", sanDomain1),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", sanDomain2),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
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

func TestAccACMCertificate_San_trailingPeriod(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	resourceName := "aws_acm_certificate.test"
	var v acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(sanDomain), acm.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile(`certificate/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          strings.TrimSuffix(sanDomain, "."),
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domain),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", strings.TrimSuffix(sanDomain, ".")),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
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

func TestAccACMCertificate_San_matches_domain(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain := rootDomain
	var v acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(sanDomain), acm.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile(`certificate/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          sanDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domain),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", sanDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
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

func TestAccACMCertificate_wildcard(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)
	var v acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(wildcardDomain, acm.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          wildcardDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
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

func TestAccACMCertificate_wildcardAndRootSan(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)
	var v acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(wildcardDomain, strconv.Quote(rootDomain), acm.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          wildcardDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
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

func TestAccACMCertificate_disableCTLogging(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	var v acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_disableCTLogging(rootDomain, acm.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.certificate_transparency_logging_preference", acm.CertificateTransparencyLoggingPreferenceDisabled),
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

// lintignore:AT002
func TestAccACMCertificate_Imported_domainName(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	commonName := "example.com"
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(caKey)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509LocallySignedCertificatePEM(t, caKey, caCertificate, key, commonName)
	newCaKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	newCaCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(newCaKey)
	newCertificate := acctest.TLSRSAX509LocallySignedCertificatePEM(t, newCaKey, newCaCertificate, key, commonName)
	withoutChainDomain := acctest.RandomDomainName()
	var v1, v2, v3 acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateKey(certificate, key, caCertificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", commonName),
				),
			},
			{
				Config: testAccCertificateConfig_privateKey(newCertificate, key, newCaCertificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v2),
					testAccCheckCertficateNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", commonName),
				),
			},
			{
				Config: testAccCertificateConfig_privateKeyNoChain(withoutChainDomain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v3),
					testAccCheckCertficateNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", withoutChainDomain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// These are not returned by the API
				ImportStateVerifyIgnore: []string{"private_key", "certificate_body", "certificate_chain"},
			},
		},
	})
}

// lintignore:AT002
func TestAccACMCertificate_Imported_validityDates(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	commonName := "example.com"
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(caKey)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509LocallySignedCertificatePEM(t, caKey, caCertificate, key, commonName)

	var v acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateKey(certificate, key, caCertificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_after"),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_before"),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", "false"),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", acm.RenewalEligibilityIneligible),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", acm.CertificateTypeImported),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// These are not returned by the API
				ImportStateVerifyIgnore: []string{"private_key", "certificate_body", "certificate_chain"},
			},
		},
	})
}

// lintignore:AT002
func TestAccACMCertificate_Imported_ipAddress(t *testing.T) { // Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7103
	resourceName := "aws_acm_certificate.test"
	var v acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateKeyNoChain("1.2.3.4"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "domain_name", ""),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// These are not returned by the API
				ImportStateVerifyIgnore: []string{"private_key", "certificate_body"},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/15055
func TestAccACMCertificate_PrivateKey_tags(t *testing.T) {
	resourceName := "aws_acm_certificate.test"
	key1 := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate1 := acctest.TLSRSAX509SelfSignedCertificatePEM(key1, "1.2.3.4")
	key2 := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate2 := acctest.TLSRSAX509SelfSignedCertificatePEM(key2, "5.6.7.8")
	var v acm.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, acm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_tags1(certificate1, key1, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_key", "certificate_body"},
			},
			{
				Config: testAccCertificateConfig_tags2(certificate1, key1, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCertificateConfig_tags1(certificate1, key1, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCertificateConfig_tags1(certificate2, key2, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckCertificateExists(n string, v *acm.CertificateDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ACM Certificate ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ACMConn

		output, err := tfacm.FindCertificateByARN(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckCertificateDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ACMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_acm_certificate" {
			continue
		}

		_, err := tfacm.FindCertificateByARN(context.Background(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("ACM Certificate %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckCertficateNotRecreated(v1, v2 *acm.CertificateDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(v1.CertificateArn) != aws.StringValue(v2.CertificateArn) {
			return fmt.Errorf("ACM Certificate recreated")
		}
		return nil
	}
}

func testAccCheckCertificateRenewed(i, j *acm.CertificateDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(j.NotAfter).After(aws.TimeValue(i.NotAfter)) {
			return fmt.Errorf("ACM Certificate not renewed: i.NotAfter=%q, j.NotAfter=%q", aws.TimeValue(i.NotAfter), aws.TimeValue(j.NotAfter))
		}

		return nil
	}
}

func testAccCheckCertificateNotRenewed(i, j *acm.CertificateDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(j.NotAfter).Equal(aws.TimeValue(i.NotAfter)) {
			return fmt.Errorf("ACM Certificate renewed: i.NotAfter=%q, j.NotAfter=%q", aws.TimeValue(i.NotAfter), aws.TimeValue(j.NotAfter))
		}

		return nil
	}
}

func testAccCertificateConfig_basic(domainName, validationMethod string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name       = %[1]q
  validation_method = %[2]q
}
`, domainName, validationMethod)
}

func testAccCertificateConfig_validationOptions(rootDomainName, domainName string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name       = %[2]q
  validation_method = "EMAIL"

  validation_option {
    domain_name       = %[2]q
    validation_domain = %[1]q
  }
}
`, rootDomainName, domainName)
}

func testAccCertificateConfig_privateCertificate_renewable(commonName, certificateDomainName string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name               = %[2]q
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  depends_on = [
    aws_acmpca_certificate_authority_certificate.test,
  ]
}

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

resource "aws_acmpca_permission" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
  principal                 = "acm.amazonaws.com"
  actions                   = ["IssueCertificate", "GetCertificate", "ListPermissions"]
}

resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 2
  }
}

resource "aws_acmpca_certificate_authority_certificate" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  certificate       = aws_acmpca_certificate.test.certificate
  certificate_chain = aws_acmpca_certificate.test.certificate_chain
}

data "aws_partition" "current" {}
`, commonName, certificateDomainName)
}

func testAccCertificateConfig_privateCertificate_noRenewalPermission(commonName, certificateDomainName string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name               = %[2]q
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  depends_on = [
    aws_acmpca_certificate_authority_certificate.test,
  ]
}

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 2
  }
}

resource "aws_acmpca_certificate_authority_certificate" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  certificate       = aws_acmpca_certificate.test.certificate
  certificate_chain = aws_acmpca_certificate.test.certificate_chain
}

data "aws_partition" "current" {}
`, commonName, certificateDomainName)
}

func testAccCertificateConfig_privateCertificate_pendingRenewal(commonName, certificateDomainName, duration string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name               = %[2]q
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  early_renewal_duration = %[3]q

  depends_on = [
    aws_acmpca_certificate_authority_certificate.test,
  ]
}

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

resource "aws_acmpca_permission" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
  principal                 = "acm.amazonaws.com"
  actions                   = ["IssueCertificate", "GetCertificate", "ListPermissions"]
}

resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 2
  }
}

resource "aws_acmpca_certificate_authority_certificate" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  certificate       = aws_acmpca_certificate.test.certificate
  certificate_chain = aws_acmpca_certificate.test.certificate_chain
}

data "aws_partition" "current" {}
`, commonName, certificateDomainName, duration)
}

func testAccCertificateConfig_subjectAlternativeNames(domainName, subjectAlternativeNames, validationMethod string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name               = %[1]q
  subject_alternative_names = [%[2]s]
  validation_method         = %[3]q
}
`, domainName, subjectAlternativeNames, validationMethod)
}

func testAccCertificateConfig_privateKeyNoChain(commonName string) string {
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(key, commonName)

	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}
`, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccCertificateConfig_tags1(certificate, key, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"

  tags = {
    %[3]q = %[4]q
  }
}
`, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), tagKey1, tagValue1)
}

func testAccCertificateConfig_tags2(certificate, key, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccCertificateConfig_privateKey(certificate, privateKey, chain string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body  = "%[1]s"
  private_key       = "%[2]s"
  certificate_chain = "%[3]s"
}
`, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(privateKey), acctest.TLSPEMEscapeNewlines(chain))
}

func testAccCertificateConfig_disableCTLogging(domainName, validationMethod string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name       = %[1]q
  validation_method = %[2]q

  options {
    certificate_transparency_logging_preference = "DISABLED"
  }
}
`, domainName, validationMethod)
}
