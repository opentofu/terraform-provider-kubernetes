// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kubernetes

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

func flattenDeploymentSpec(in appsv1.DeploymentSpec, d *schema.ResourceData, meta interface{}) ([]interface{}, error) {
	att := make(map[string]interface{})
	att["min_ready_seconds"] = in.MinReadySeconds

	if in.Replicas != nil {
		att["replicas"] = strconv.Itoa(int(*in.Replicas))
	}

	if in.ProgressDeadlineSeconds != nil {
		att["progress_deadline_seconds"] = *in.ProgressDeadlineSeconds
	}

	if in.RevisionHistoryLimit != nil {
		att["revision_history_limit"] = *in.RevisionHistoryLimit
	}

	if in.Selector != nil {
		att["selector"] = flattenLabelSelector(in.Selector)
	}

	att["strategy"] = flattenDeploymentStrategy(in.Strategy)

	podSpec, err := flattenPodSpec(in.Template.Spec, true)
	if err != nil {
		return nil, err
	}
	template := make(map[string]interface{})
	template["spec"] = podSpec
	template["metadata"] = flattenMetadataFields(in.Template.ObjectMeta)
	att["template"] = []interface{}{template}

	return []interface{}{att}, nil
}

func flattenDeploymentStrategy(in appsv1.DeploymentStrategy) []interface{} {
	att := make(map[string]interface{})
	if in.Type != "" {
		att["type"] = in.Type
	}
	if in.RollingUpdate != nil {
		att["rolling_update"] = flattenDeploymentStrategyRollingUpdate(in.RollingUpdate)
	}
	return []interface{}{att}
}

func flattenDeploymentStrategyRollingUpdate(in *appsv1.RollingUpdateDeployment) []interface{} {
	att := make(map[string]interface{})
	if in.MaxUnavailable != nil {
		att["max_unavailable"] = in.MaxUnavailable.String()
	}
	if in.MaxSurge != nil {
		att["max_surge"] = in.MaxSurge.String()
	}
	return []interface{}{att}
}

func expandDeploymentSpec(deployment []interface{}) (*appsv1.DeploymentSpec, error) {
	obj := &appsv1.DeploymentSpec{}

	if len(deployment) == 0 || deployment[0] == nil {
		return obj, nil
	}

	in := deployment[0].(map[string]interface{})

	obj.MinReadySeconds = int32(in["min_ready_seconds"].(int))
	obj.Paused = in["paused"].(bool)
	obj.ProgressDeadlineSeconds = ptr.To(int32(in["progress_deadline_seconds"].(int)))
	if v, ok := in["replicas"].(string); ok && v != "" {
		i, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return obj, err
		}
		obj.Replicas = ptr.To(int32(i))
	}
	obj.RevisionHistoryLimit = ptr.To(int32(in["revision_history_limit"].(int)))

	if v, ok := in["selector"].([]interface{}); ok && len(v) > 0 {
		obj.Selector = expandLabelSelector(v)
	}

	if v, ok := in["strategy"].([]interface{}); ok && len(v) > 0 {
		obj.Strategy = expandDeploymentStrategy(v)
	}

	template, err := expandPodTemplate(in["template"].([]interface{}))
	if err != nil {
		return obj, err
	}
	obj.Template = *template

	return obj, nil
}

func expandPodTemplate(l []interface{}) (*corev1.PodTemplateSpec, error) {
	obj := &corev1.PodTemplateSpec{}
	if len(l) == 0 || l[0] == nil {
		return obj, nil
	}
	in := l[0].(map[string]interface{})

	obj.ObjectMeta = expandMetadata(in["metadata"].([]interface{}))

	if v, ok := in["spec"].([]interface{}); ok && len(v) > 0 {
		podSpec, err := expandPodSpec(in["spec"].([]interface{}))
		if err != nil {
			return obj, err
		}
		obj.Spec = *podSpec
	}
	return obj, nil
}

func expandDeploymentStrategy(l []interface{}) appsv1.DeploymentStrategy {
	obj := appsv1.DeploymentStrategy{}

	if len(l) == 0 || l[0] == nil {
		obj.Type = appsv1.RollingUpdateDeploymentStrategyType
		return obj
	}
	in := l[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok {
		obj.Type = appsv1.DeploymentStrategyType(v)
	}
	if v, ok := in["rolling_update"].([]interface{}); ok && len(v) > 0 {
		obj.RollingUpdate = expandRollingUpdateDeployment(v)
	}
	return obj
}

func expandRollingUpdateDeployment(l []interface{}) *appsv1.RollingUpdateDeployment {
	obj := appsv1.RollingUpdateDeployment{}
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	in := l[0].(map[string]interface{})

	if v, ok := in["max_surge"].(string); ok {
		val := intstr.Parse(v)
		obj.MaxSurge = &val
	}
	if v, ok := in["max_unavailable"].(string); ok {
		val := intstr.Parse(v)
		obj.MaxUnavailable = &val
	}
	return &obj
}
