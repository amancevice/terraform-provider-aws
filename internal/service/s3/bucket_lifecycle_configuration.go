// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/types/nullable"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"golang.org/x/exp/slices"
)

// @SDKResource("aws_s3_bucket_lifecycle_configuration")
func ResourceBucketLifecycleConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketLifecycleConfigurationCreate,
		ReadWithoutTimeout:   resourceBucketLifecycleConfigurationRead,
		UpdateWithoutTimeout: resourceBucketLifecycleConfigurationUpdate,
		DeleteWithoutTimeout: resourceBucketLifecycleConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 63),
			},
			"expected_bucket_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"rule": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"abort_incomplete_multipart_upload": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days_after_initiation": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"expiration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"date": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidUTCTimestamp,
									},
									"days": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  0, // API returns 0
									},
									"expired_object_delete_marker": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true, // API returns false; conflicts with date and days
									},
								},
							},
						},
						"filter": {
							Type:     schema.TypeList,
							Optional: true,
							// If neither the filter block nor the prefix parameter in the rule are specified,
							// we apply the Default behavior from v3.x of the provider (Filter with empty string Prefix),
							// which will thus return a Filter in the GetBucketLifecycleConfiguration request and
							// require diff suppression.
							DiffSuppressFunc: suppressMissingFilterConfigurationBlock,
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"and": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object_size_greater_than": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(0),
												},
												"object_size_less_than": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
												"prefix": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"tags": tftags.TagsSchema(),
											},
										},
									},
									"object_size_greater_than": {
										Type:     nullable.TypeNullableInt,
										Optional: true,
									},
									"object_size_less_than": {
										Type:     nullable.TypeNullableInt,
										Optional: true,
									},
									"prefix": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"tag": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:     schema.TypeString,
													Required: true,
												},
												"value": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"noncurrent_version_expiration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"newer_noncurrent_versions": {
										Type:         nullable.TypeNullableInt,
										Optional:     true,
										ValidateFunc: nullable.ValidateTypeStringNullableIntAtLeast(1),
									},
									"noncurrent_days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"noncurrent_version_transition": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"newer_noncurrent_versions": {
										Type:         nullable.TypeNullableInt,
										Optional:     true,
										ValidateFunc: nullable.ValidateTypeStringNullableIntAtLeast(1),
									},
									"noncurrent_days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"storage_class": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.TransitionStorageClass](),
									},
								},
							},
						},
						"prefix": {
							Type:       schema.TypeString,
							Optional:   true,
							Deprecated: "Use filter instead",
						},
						"status": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(lifecycleRuleStatus_Values(), false),
						},
						"transition": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"date": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidUTCTimestamp,
									},
									"days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"storage_class": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.TransitionStorageClass](),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceBucketLifecycleConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Get("bucket").(string)
	expectedBucketOwner := d.Get("expected_bucket_owner").(string)
	rules := expandLifecycleRules(ctx, d.Get("rule").([]interface{}))
	input := &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucket),
		LifecycleConfiguration: &types.BucketLifecycleConfiguration{
			Rules: rules,
		},
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, s3BucketPropagationTimeout, func() (interface{}, error) {
		return conn.PutBucketLifecycleConfiguration(ctx, input)
	}, errCodeNoSuchBucket)

	if tfawserr.ErrMessageContains(err, errCodeInvalidArgument, "LifecycleConfiguration is not valid, expected CreateBucketConfiguration") {
		err = errDirectoryBucket(err)
	}

	if err != nil {
		return diag.Errorf("creating S3 Bucket (%s) Lifecycle Configuration: %s", bucket, err)
	}

	d.SetId(CreateResourceID(bucket, expectedBucketOwner))

	_, err = waitLifecycleRulesEquals(ctx, conn, bucket, expectedBucketOwner, rules, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		diag.Errorf("waiting for S3 Bucket Lifecycle Configuration (%s) create: %s", d.Id(), err)
	}

	return resourceBucketLifecycleConfigurationRead(ctx, d, meta)
}

func resourceBucketLifecycleConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	const (
		lifecycleConfigurationExtraRetryDelay    = 5 * time.Second
		lifecycleConfigurationRulesSteadyTimeout = 2 * time.Minute
	)
	var lastOutput, output []types.LifecycleRule

	err = retry.RetryContext(ctx, lifecycleConfigurationRulesSteadyTimeout, func() *retry.RetryError {
		var err error

		time.Sleep(lifecycleConfigurationExtraRetryDelay)

		output, err = findLifecycleRules(ctx, conn, bucket, expectedBucketOwner)

		if d.IsNewResource() && tfresource.NotFound(err) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		if lastOutput == nil || !lifecycleRulesEqual(lastOutput, output) {
			lastOutput = output
			return retry.RetryableError(fmt.Errorf("S3 Bucket Lifecycle Configuration (%s) has not stablized; retrying", d.Id()))
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = findLifecycleRules(ctx, conn, bucket, expectedBucketOwner)
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Lifecycle Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading S3 Bucket Lifecycle Configuration (%s): %s", d.Id(), err)
	}

	d.Set("bucket", bucket)
	d.Set("expected_bucket_owner", expectedBucketOwner)
	if err := d.Set("rule", flattenLifecycleRules(ctx, output)); err != nil {
		return diag.Errorf("setting rule: %s", err)
	}

	return nil
}

func resourceBucketLifecycleConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	rules := expandLifecycleRules(ctx, d.Get("rule").([]interface{}))
	input := &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucket),
		LifecycleConfiguration: &types.BucketLifecycleConfiguration{
			Rules: rules,
		},
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, s3BucketPropagationTimeout, func() (interface{}, error) {
		return conn.PutBucketLifecycleConfiguration(ctx, input)
	}, errCodeNoSuchLifecycleConfiguration)

	if err != nil {
		return diag.Errorf("updating S3 Bucket Lifecycle Configuration (%s): %s", d.Id(), err)
	}

	_, err = waitLifecycleRulesEquals(ctx, conn, bucket, expectedBucketOwner, rules, d.Timeout(schema.TimeoutUpdate))

	if err != nil {
		diag.Errorf("waiting for S3 Bucket Lifecycle Configuration (%s) update: %s", d.Id(), err)
	}

	return resourceBucketLifecycleConfigurationRead(ctx, d, meta)
}

func resourceBucketLifecycleConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.DeleteBucketLifecycleInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.DeleteBucketLifecycle(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchLifecycleConfiguration) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting S3 Bucket Lifecycle Configuration (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, s3BucketPropagationTimeout, func() (interface{}, error) {
		return findLifecycleRules(ctx, conn, bucket, expectedBucketOwner)
	})

	if err != nil {
		return diag.Errorf("waiting for S3 Bucket Lifecyle Configuration (%s) delete: %s", d.Id(), err)
	}

	return nil
}

// suppressMissingFilterConfigurationBlock suppresses the diff that results from an omitted
// filter configuration block and one returned from the S3 API.
// To work around the issue, https://github.com/hashicorp/terraform-plugin-sdk/issues/743,
// this method only looks for changes in the "filter.#" value and not its nested fields
// which are incorrectly suppressed when using the verify.SuppressMissingOptionalConfigurationBlock method.
func suppressMissingFilterConfigurationBlock(k, old, new string, d *schema.ResourceData) bool {
	if strings.HasSuffix(k, "filter.#") {
		oraw, nraw := d.GetChange(k)
		o, n := oraw.(int), nraw.(int)

		if o == 1 && n == 0 {
			return true
		}

		if o == 1 && n == 1 {
			return old == "1" && new == "0"
		}

		return false
	}
	return false
}

func findLifecycleRules(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string) ([]types.LifecycleRule, error) {
	input := &s3.GetBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetBucketLifecycleConfiguration(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchLifecycleConfiguration) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Rules) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Rules, nil
}

func lifecycleRulesEqual(rules1, rules2 []types.LifecycleRule) bool {
	if len(rules1) != len(rules2) {
		return false
	}

	for _, rule1 := range rules1 {
		// We consider 2 LifecycleRules equal if their IDs and Statuses are equal.
		if !slices.ContainsFunc(rules2, func(rule2 types.LifecycleRule) bool {
			return aws.ToString(rule1.ID) == aws.ToString(rule2.ID) && rule1.Status == rule2.Status
		}) {
			return false
		}
	}

	return true
}

func statusLifecycleRulesEquals(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string, rules []types.LifecycleRule) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findLifecycleRules(ctx, conn, bucket, expectedBucketOwner)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(lifecycleRulesEqual(output, rules)), nil
	}
}

func waitLifecycleRulesEquals(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string, rules []types.LifecycleRule, timeout time.Duration) ([]types.LifecycleRule, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Target:                    []string{strconv.FormatBool(true)},
		Refresh:                   statusLifecycleRulesEquals(ctx, conn, bucket, expectedBucketOwner, rules),
		Timeout:                   timeout,
		MinTimeout:                10 * time.Second,
		ContinuousTargetOccurence: 3,
		NotFoundChecks:            20,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]types.LifecycleRule); ok {
		return output, err
	}

	return nil, err
}

const (
	lifecycleRuleStatusDisabled = "Disabled"
	lifecycleRuleStatusEnabled  = "Enabled"
)

func lifecycleRuleStatus_Values() []string {
	return []string{
		lifecycleRuleStatusDisabled,
		lifecycleRuleStatusEnabled,
	}
}

func expandLifecycleRules(ctx context.Context, l []interface{}) []types.LifecycleRule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var results []types.LifecycleRule

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		result := types.LifecycleRule{}

		if v, ok := tfMap["abort_incomplete_multipart_upload"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			result.AbortIncompleteMultipartUpload = expandLifecycleRuleAbortIncompleteMultipartUpload(v[0].(map[string]interface{}))
		}

		if v, ok := tfMap["expiration"].([]interface{}); ok && len(v) > 0 {
			result.Expiration = expandLifecycleRuleExpiration(v)
		}

		if v, ok := tfMap["filter"].([]interface{}); ok && len(v) > 0 {
			result.Filter = expandLifecycleRuleFilter(ctx, v)
		}

		if v, ok := tfMap["prefix"].(string); ok && result.Filter == nil {
			// If neither the filter block nor the prefix are specified,
			// apply the Default behavior from v3.x of the provider;
			// otherwise, set the prefix as specified in Terraform.
			if v == "" {
				result.Filter = &types.LifecycleRuleFilterMemberPrefix{
					Value: v,
				}
			} else {
				result.Prefix = aws.String(v)
			}
		}

		if v, ok := tfMap["id"].(string); ok {
			result.ID = aws.String(v)
		}

		if v, ok := tfMap["noncurrent_version_expiration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			result.NoncurrentVersionExpiration = expandLifecycleRuleNoncurrentVersionExpiration(v[0].(map[string]interface{}))
		}

		if v, ok := tfMap["noncurrent_version_transition"].(*schema.Set); ok && v.Len() > 0 {
			result.NoncurrentVersionTransitions = expandLifecycleRuleNoncurrentVersionTransitions(v.List())
		}

		if v, ok := tfMap["status"].(string); ok && v != "" {
			result.Status = types.ExpirationStatus(v)
		}

		if v, ok := tfMap["transition"].(*schema.Set); ok && v.Len() > 0 {
			result.Transitions = expandLifecycleRuleTransitions(v.List())
		}

		results = append(results, result)
	}

	return results
}

func expandLifecycleRuleAbortIncompleteMultipartUpload(m map[string]interface{}) *types.AbortIncompleteMultipartUpload {
	if len(m) == 0 {
		return nil
	}

	result := &types.AbortIncompleteMultipartUpload{}

	if v, ok := m["days_after_initiation"].(int); ok {
		result.DaysAfterInitiation = aws.Int32(int32(v))
	}

	return result
}

func expandLifecycleRuleExpiration(l []interface{}) *types.LifecycleExpiration {
	if len(l) == 0 {
		return nil
	}

	result := &types.LifecycleExpiration{}

	if l[0] == nil {
		return result
	}

	m := l[0].(map[string]interface{})

	if v, ok := m["date"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v)
		result.Date = aws.Time(t)
	}

	if v, ok := m["days"].(int); ok && v > 0 {
		result.Days = aws.Int32(int32(v))
	}

	// This cannot be specified with Days or Date
	if v, ok := m["expired_object_delete_marker"].(bool); ok && result.Date == nil && aws.ToInt32(result.Days) == 0 {
		result.ExpiredObjectDeleteMarker = aws.Bool(v)
	}

	return result
}

func expandLifecycleRuleFilter(ctx context.Context, l []interface{}) types.LifecycleRuleFilter {
	if len(l) == 0 {
		return nil
	}

	var result types.LifecycleRuleFilter

	if l[0] == nil {
		return result
	}

	m := l[0].(map[string]interface{})

	if v, ok := m["and"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result = expandLifecycleRuleFilterMemberAnd(ctx, v[0].(map[string]interface{}))
	}

	if v, null, _ := nullable.Int(m["object_size_greater_than"].(string)).Value(); !null && v >= 0 {
		result = &types.LifecycleRuleFilterMemberObjectSizeGreaterThan{
			Value: v,
		}
	}

	if v, null, _ := nullable.Int(m["object_size_less_than"].(string)).Value(); !null && v > 0 {
		result = &types.LifecycleRuleFilterMemberObjectSizeLessThan{
			Value: v,
		}
	}

	if v, ok := m["tag"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result = expandLifecycleRuleFilterMemberTag(v[0].(map[string]interface{}))
	}

	// Per AWS S3 API, "A Filter must have exactly one of Prefix, Tag, or And specified";
	// Specifying more than one of the listed parameters results in a MalformedXML error.
	// In practice, this also includes ObjectSizeGreaterThan and ObjectSizeLessThan.
	if v, ok := m["prefix"].(string); ok && result == nil {
		result = &types.LifecycleRuleFilterMemberPrefix{
			Value: v,
		}
	}

	return result
}

func expandLifecycleRuleFilterMemberAnd(ctx context.Context, m map[string]interface{}) *types.LifecycleRuleFilterMemberAnd {
	if len(m) == 0 {
		return nil
	}

	result := &types.LifecycleRuleFilterMemberAnd{
		Value: types.LifecycleRuleAndOperator{},
	}

	if v, ok := m["object_size_greater_than"].(int); ok && v > 0 {
		result.Value.ObjectSizeGreaterThan = aws.Int64(int64(v))
	}

	if v, ok := m["object_size_less_than"].(int); ok && v > 0 {
		result.Value.ObjectSizeLessThan = aws.Int64(int64(v))
	}

	if v, ok := m["prefix"].(string); ok {
		result.Value.Prefix = aws.String(v)
	}

	if v, ok := m["tags"].(map[string]interface{}); ok && len(v) > 0 {
		tags := Tags(tftags.New(ctx, v).IgnoreAWS())
		if len(tags) > 0 {
			result.Value.Tags = tags
		}
	}

	return result
}

func expandLifecycleRuleFilterMemberTag(m map[string]interface{}) *types.LifecycleRuleFilterMemberTag {
	if len(m) == 0 {
		return nil
	}

	result := &types.LifecycleRuleFilterMemberTag{
		Value: types.Tag{},
	}

	if key, ok := m["key"].(string); ok {
		result.Value.Key = aws.String(key)
	}

	if value, ok := m["value"].(string); ok {
		result.Value.Value = aws.String(value)
	}

	return result
}

func expandLifecycleRuleNoncurrentVersionExpiration(m map[string]interface{}) *types.NoncurrentVersionExpiration {
	if len(m) == 0 {
		return nil
	}

	result := &types.NoncurrentVersionExpiration{}

	if v, null, _ := nullable.Int(m["newer_noncurrent_versions"].(string)).Value(); !null && v > 0 {
		result.NewerNoncurrentVersions = aws.Int32(int32(v))
	}

	if v, ok := m["noncurrent_days"].(int); ok {
		result.NoncurrentDays = aws.Int32(int32(v))
	}

	return result
}

func expandLifecycleRuleNoncurrentVersionTransitions(l []interface{}) []types.NoncurrentVersionTransition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var results []types.NoncurrentVersionTransition

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		transition := types.NoncurrentVersionTransition{}

		if v, null, _ := nullable.Int(tfMap["newer_noncurrent_versions"].(string)).Value(); !null && v > 0 {
			transition.NewerNoncurrentVersions = aws.Int32(int32(v))
		}

		if v, ok := tfMap["noncurrent_days"].(int); ok {
			transition.NoncurrentDays = aws.Int32(int32(v))
		}

		if v, ok := tfMap["storage_class"].(string); ok && v != "" {
			transition.StorageClass = types.TransitionStorageClass(v)
		}

		results = append(results, transition)
	}

	return results
}

func expandLifecycleRuleTransitions(l []interface{}) []types.Transition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var results []types.Transition

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		transition := types.Transition{}

		if v, ok := tfMap["date"].(string); ok && v != "" {
			t, _ := time.Parse(time.RFC3339, v)
			transition.Date = aws.Time(t)
		}

		// Only one of "date" and "days" can be configured
		// so only set the transition.Days value when transition.Date is nil
		// By default, tfMap["days"] = 0 if not explicitly configured in terraform.
		if v, ok := tfMap["days"].(int); ok && v >= 0 && transition.Date == nil {
			transition.Days = aws.Int32(int32(v))
		}

		if v, ok := tfMap["storage_class"].(string); ok && v != "" {
			transition.StorageClass = types.TransitionStorageClass(v)
		}

		results = append(results, transition)
	}

	return results
}

func flattenLifecycleRules(ctx context.Context, rules []types.LifecycleRule) []interface{} {
	if len(rules) == 0 {
		return []interface{}{}
	}

	var results []interface{}

	for _, rule := range rules {
		m := map[string]interface{}{
			"status": rule.Status,
		}

		if rule.AbortIncompleteMultipartUpload != nil {
			m["abort_incomplete_multipart_upload"] = flattenLifecycleRuleAbortIncompleteMultipartUpload(rule.AbortIncompleteMultipartUpload)
		}

		if rule.Expiration != nil {
			m["expiration"] = flattenLifecycleRuleExpiration(rule.Expiration)
		}

		if rule.Filter != nil {
			m["filter"] = flattenLifecycleRuleFilter(ctx, rule.Filter)
		}

		if rule.ID != nil {
			m["id"] = aws.ToString(rule.ID)
		}

		if rule.NoncurrentVersionExpiration != nil {
			m["noncurrent_version_expiration"] = flattenLifecycleRuleNoncurrentVersionExpiration(rule.NoncurrentVersionExpiration)
		}

		if rule.NoncurrentVersionTransitions != nil {
			m["noncurrent_version_transition"] = flattenLifecycleRuleNoncurrentVersionTransitions(rule.NoncurrentVersionTransitions)
		}

		if rule.Prefix != nil {
			m["prefix"] = aws.ToString(rule.Prefix)
		}

		if rule.Transitions != nil {
			m["transition"] = flattenLifecycleRuleTransitions(rule.Transitions)
		}

		results = append(results, m)
	}

	return results
}

func flattenLifecycleRuleAbortIncompleteMultipartUpload(u *types.AbortIncompleteMultipartUpload) []interface{} {
	if u == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if u.DaysAfterInitiation != nil {
		m["days_after_initiation"] = int(aws.ToInt32(u.DaysAfterInitiation))
	}

	return []interface{}{m}
}

func flattenLifecycleRuleExpiration(expiration *types.LifecycleExpiration) []interface{} {
	if expiration == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if expiration.Date != nil {
		m["date"] = expiration.Date.Format(time.RFC3339)
	}

	if expiration.Days != nil {
		m["days"] = int(aws.ToInt32(expiration.Days))
	}

	if expiration.ExpiredObjectDeleteMarker != nil {
		m["expired_object_delete_marker"] = aws.ToBool(expiration.ExpiredObjectDeleteMarker)
	}

	return []interface{}{m}
}

func flattenLifecycleRuleFilter(ctx context.Context, filter types.LifecycleRuleFilter) []interface{} {
	if filter == nil {
		return nil
	}

	m := make(map[string]interface{})

	switch v := filter.(type) {
	case *types.LifecycleRuleFilterMemberAnd:
		m["and"] = flattenLifecycleRuleFilterMemberAnd(ctx, v)
	case *types.LifecycleRuleFilterMemberObjectSizeGreaterThan:
		m["object_size_greater_than"] = strconv.FormatInt(v.Value, 10)
	case *types.LifecycleRuleFilterMemberObjectSizeLessThan:
		m["object_size_less_than"] = strconv.FormatInt(v.Value, 10)
	case *types.LifecycleRuleFilterMemberPrefix:
		m["prefix"] = v.Value
	case *types.LifecycleRuleFilterMemberTag:
		m["tag"] = flattenLifecycleRuleFilterMemberTag(v)
	default:
		return nil
	}

	return []interface{}{m}
}

func flattenLifecycleRuleFilterMemberAnd(ctx context.Context, andOp *types.LifecycleRuleFilterMemberAnd) []interface{} {
	if andOp == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"object_size_greater_than": andOp.Value.ObjectSizeGreaterThan,
		"object_size_less_than":    andOp.Value.ObjectSizeLessThan,
	}

	if v := andOp.Value.Prefix; v != nil {
		m["prefix"] = aws.ToString(v)
	}

	if v := andOp.Value.Tags; v != nil {
		m["tags"] = keyValueTags(ctx, v).IgnoreAWS().Map()
	}

	return []interface{}{m}
}

func flattenLifecycleRuleFilterMemberTag(op *types.LifecycleRuleFilterMemberTag) []interface{} {
	if op == nil {
		return nil
	}

	m := make(map[string]interface{})

	if v := op.Value.Key; v != nil {
		m["key"] = aws.ToString(v)
	}

	if v := op.Value.Value; v != nil {
		m["value"] = aws.ToString(v)
	}

	return []interface{}{m}
}

func flattenLifecycleRuleNoncurrentVersionExpiration(expiration *types.NoncurrentVersionExpiration) []interface{} {
	if expiration == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if expiration.NewerNoncurrentVersions != nil {
		m["newer_noncurrent_versions"] = strconv.FormatInt(int64(aws.ToInt32(expiration.NewerNoncurrentVersions)), 10)
	}

	if expiration.NoncurrentDays != nil {
		m["noncurrent_days"] = int(aws.ToInt32(expiration.NoncurrentDays))
	}

	return []interface{}{m}
}

func flattenLifecycleRuleNoncurrentVersionTransitions(transitions []types.NoncurrentVersionTransition) []interface{} {
	if len(transitions) == 0 {
		return []interface{}{}
	}

	var results []interface{}

	for _, transition := range transitions {
		m := map[string]interface{}{
			"storage_class": transition.StorageClass,
		}

		if transition.NewerNoncurrentVersions != nil {
			m["newer_noncurrent_versions"] = strconv.FormatInt(int64(aws.ToInt32(transition.NewerNoncurrentVersions)), 10)
		}

		if transition.NoncurrentDays != nil {
			m["noncurrent_days"] = int(aws.ToInt32(transition.NoncurrentDays))
		}

		results = append(results, m)
	}

	return results
}

func flattenLifecycleRuleTransitions(transitions []types.Transition) []interface{} {
	if len(transitions) == 0 {
		return []interface{}{}
	}

	var results []interface{}

	for _, transition := range transitions {
		m := map[string]interface{}{
			"days":          transition.Days,
			"storage_class": transition.StorageClass,
		}

		if transition.Date != nil {
			m["date"] = transition.Date.Format(time.RFC3339)
		}

		results = append(results, m)
	}

	return results
}
