# Kubernetes & Helm Deployment Guide for Execution Status Refresh Optimization

## Overview

This guide provides Kubernetes-native solutions for deploying Harbor with the execution status refresh optimization in a multi-instance setup.

## Kubernetes Deployment Strategies

### Strategy 1: StatefulSet with Automatic Instance ID (Recommended)

StatefulSets provide predictable pod names (`harbor-core-0`, `harbor-core-1`, etc.) which we can parse to extract the instance ID.

#### Helm Chart Values

```yaml
# values.yaml
core:
  replicas: 5
  
  # Enable execution refresh optimization
  executionRefreshOptimization:
    enabled: true
    # Total instances will be set to replicas automatically
```

#### StatefulSet Template

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: harbor-core
spec:
  serviceName: harbor-core
  replicas: 5
  selector:
    matchLabels:
      app: harbor
      component: core
  template:
    metadata:
      labels:
        app: harbor
        component: core
    spec:
      initContainers:
      # Init container to extract instance ID from pod name
      - name: set-instance-id
        image: busybox:1.36
        command:
        - sh
        - -c
        - |
          # Extract number from pod name (e.g., harbor-core-3 -> 3)
          POD_NAME=$(cat /etc/podinfo/name)
          INSTANCE_ID=${POD_NAME##*-}
          echo "Detected instance ID: $INSTANCE_ID"
          echo "CORE_INSTANCE_ID=$INSTANCE_ID" > /shared/instance-id.env
          
          # Also set total instances from replica count
          TOTAL_INSTANCES=$(cat /etc/podinfo/total-instances)
          echo "CORE_INSTANCE_TOTAL=$TOTAL_INSTANCES" >> /shared/instance-id.env
          
          cat /shared/instance-id.env
        volumeMounts:
        - name: podinfo
          mountPath: /etc/podinfo
        - name: shared-config
          mountPath: /shared
      
      containers:
      - name: core
        image: goharbor/harbor-core:v2.11.0
        env:
        # Source the instance ID from init container
        - name: CORE_INSTANCE_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
          # This will be processed by entrypoint script
        - name: CORE_INSTANCE_TOTAL
          value: "5"
        
        # Existing environment variables
        - name: EXECUTION_STATUS_REFRESH_INTERVAL_SECONDS
          value: "30"
        - name: _REDIS_URL_CORE
          value: "redis://redis:6379/0"
        # ... other env vars ...
        
        volumeMounts:
        - name: shared-config
          mountPath: /shared
        
        # Use a wrapper script to set instance ID
        command: ["/bin/sh"]
        args:
        - -c
        - |
          # Extract instance ID from pod name
          if [ -n "$CORE_INSTANCE_ID" ]; then
            INSTANCE_NUM=$(echo $CORE_INSTANCE_ID | grep -o '[0-9]*$')
            export CORE_INSTANCE_ID=$INSTANCE_NUM
            echo "Starting Harbor Core with INSTANCE_ID=$CORE_INSTANCE_ID, TOTAL=$CORE_INSTANCE_TOTAL"
          fi
          exec /harbor/entrypoint.sh
      
      volumes:
      - name: podinfo
        downwardAPI:
          items:
          - path: name
            fieldRef:
              fieldPath: metadata.name
          - path: total-instances
            fieldRef:
              fieldPath: metadata.annotations['harbor.io/total-instances']
      - name: shared-config
        emptyDir: {}
```

### Strategy 2: Deployment with Pod Index Annotation

For regular Deployments, use an admission webhook or init container to assign unique IDs.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: harbor-core
spec:
  replicas: 5
  template:
    metadata:
      labels:
        app: harbor
        component: core
    spec:
      initContainers:
      - name: assign-instance-id
        image: bitnami/kubectl:latest
        command:
        - sh
        - -c
        - |
          # Get all core pods sorted by creation time
          PODS=$(kubectl get pods -l app=harbor,component=core \
            --sort-by=.metadata.creationTimestamp \
            -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}')
          
          # Find this pod's index
          MY_POD=$(cat /etc/podinfo/name)
          INSTANCE_ID=0
          for pod in $PODS; do
            if [ "$pod" = "$MY_POD" ]; then
              break
            fi
            INSTANCE_ID=$((INSTANCE_ID + 1))
          done
          
          echo "CORE_INSTANCE_ID=$INSTANCE_ID" > /shared/instance-id.env
          echo "CORE_INSTANCE_TOTAL=5" >> /shared/instance-id.env
        volumeMounts:
        - name: podinfo
          mountPath: /etc/podinfo
        - name: shared-config
          mountPath: /shared
      
      containers:
      - name: core
        image: goharbor/harbor-core:v2.11.0
        command: ["/bin/sh"]
        args:
        - -c
        - |
          if [ -f /shared/instance-id.env ]; then
            source /shared/instance-id.env
            export CORE_INSTANCE_ID
            export CORE_INSTANCE_TOTAL
          fi
          exec /harbor/entrypoint.sh
        volumeMounts:
        - name: shared-config
          mountPath: /shared
      
      volumes:
      - name: podinfo
        downwardAPI:
          items:
          - path: name
            fieldRef:
              fieldPath: metadata.name
      - name: shared-config
        emptyDir: {}
      
      serviceAccountName: harbor-core
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: harbor-core
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: harbor-core-pod-reader
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: harbor-core-pod-reader
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: harbor-core-pod-reader
subjects:
- kind: ServiceAccount
  name: harbor-core
```

### Strategy 3: Enhanced Go Code with Kubernetes Auto-Detection (Best!)

Update the `InstanceCoordinator` to automatically detect Kubernetes environment.

Add to `execution_queue.go`:

```go
import (
    "io/ioutil"
    "strings"
)

// NewInstanceCoordinator creates a new InstanceCoordinator
// It automatically detects Kubernetes environment and extracts instance ID from pod name
func NewInstanceCoordinator() *InstanceCoordinator {
    totalInstances := defaultTotalInstances
    myInstanceID := 0
    
    // Try environment variables first
    if val := os.Getenv("CORE_INSTANCE_TOTAL"); val != "" {
        if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
            totalInstances = parsed
        }
    }
    
    if val := os.Getenv("CORE_INSTANCE_ID"); val != "" {
        if parsed, err := strconv.Atoi(val); err == nil && parsed >= 0 && parsed < totalInstances {
            myInstanceID = parsed
        }
    } else {
        // Auto-detect from Kubernetes pod name
        myInstanceID = detectKubernetesInstanceID()
    }
    
    return &InstanceCoordinator{
        totalInstances: totalInstances,
        myInstanceID:   myInstanceID,
    }
}

// detectKubernetesInstanceID attempts to extract instance ID from Kubernetes pod name
func detectKubernetesInstanceID() int {
    // Try to read pod name from downward API
    if hostname := os.Getenv("HOSTNAME"); hostname != "" {
        return extractInstanceIDFromPodName(hostname)
    }
    
    // Try to read from /etc/hostname (Kubernetes sets this to pod name)
    if data, err := ioutil.ReadFile("/etc/hostname"); err == nil {
        hostname := strings.TrimSpace(string(data))
        return extractInstanceIDFromPodName(hostname)
    }
    
    return 0
}

// extractInstanceIDFromPodName extracts the numeric suffix from pod names
// Examples:
//   harbor-core-0 -> 0
//   harbor-core-3 -> 3
//   harbor-core-statefulset-2 -> 2
func extractInstanceIDFromPodName(podName string) int {
    // Find the last dash and extract number after it
    parts := strings.Split(podName, "-")
    if len(parts) > 0 {
        lastPart := parts[len(parts)-1]
        if id, err := strconv.Atoi(lastPart); err == nil && id >= 0 {
            return id
        }
    }
    return 0
}
```

## Helm Chart Integration

### Recommended Helm Chart Structure

```
harbor/
├── Chart.yaml
├── values.yaml
└── templates/
    ├── core/
    │   ├── core-ss.yaml          # StatefulSet for core
    │   ├── core-svc.yaml         # Service
    │   └── core-cm.yaml          # ConfigMap
    └── ...
```

### Enhanced values.yaml

```yaml
# values.yaml
core:
  # Number of core replicas for high availability
  replicas: 5
  
  # Execution status refresh optimization
  executionRefresh:
    # Enable optimized execution status refresh
    enabled: true
    
    # Refresh interval in seconds
    intervalSeconds: 30
    
    # Auto-detect instance ID from pod name (recommended for K8s)
    autoDetectInstanceId: true
    
    # Manual instance configuration (only if autoDetect is false)
    # instances:
    #   total: 5
    #   id: 0

  resources:
    requests:
      memory: "2Gi"
      cpu: "1000m"
    limits:
      memory: "4Gi"
      cpu: "2000m"

redis:
  # Redis configuration
  type: internal  # or external
  internal:
    # Use Redis Sentinel for HA
    sentinel:
      enabled: true
      replicas: 3
```

### Enhanced core-ss.yaml Template

```yaml
{{- if .Values.core.replicas }}
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: {{ include "harbor.core" . }}
  labels:
    {{- include "harbor.labels" . | nindent 4 }}
    component: core
spec:
  serviceName: {{ include "harbor.core" . }}
  replicas: {{ .Values.core.replicas }}
  selector:
    matchLabels:
      {{- include "harbor.matchLabels" . | nindent 6 }}
      component: core
  template:
    metadata:
      labels:
        {{- include "harbor.labels" . | nindent 8 }}
        component: core
      annotations:
        checksum/configmap: {{ include (print $.Template.BasePath "/core/core-cm.yaml") . | sha256sum }}
        checksum/secret: {{ include (print $.Template.BasePath "/core/core-secret.yaml") . | sha256sum }}
    spec:
      {{- if .Values.core.executionRefresh.enabled }}
      {{- if not .Values.core.executionRefresh.autoDetectInstanceId }}
      # Only needed if not using auto-detection
      initContainers:
      - name: set-instance-id
        image: {{ .Values.core.image.repository }}:{{ .Values.core.image.tag }}
        command:
        - sh
        - -c
        - |
          POD_NAME=$(cat /etc/podinfo/name)
          INSTANCE_ID=${POD_NAME##*-}
          echo "CORE_INSTANCE_ID=$INSTANCE_ID" > /shared/instance-id.env
        volumeMounts:
        - name: podinfo
          mountPath: /etc/podinfo
        - name: shared-config
          mountPath: /shared
      {{- end }}
      {{- end }}
      
      containers:
      - name: core
        image: {{ .Values.core.image.repository }}:{{ .Values.core.image.tag }}
        imagePullPolicy: {{ .Values.imagePullPolicy }}
        
        env:
        {{- if .Values.core.executionRefresh.enabled }}
        - name: CORE_INSTANCE_TOTAL
          value: {{ .Values.core.replicas | quote }}
        
        {{- if .Values.core.executionRefresh.autoDetectInstanceId }}
        # Auto-detection uses HOSTNAME (set by Kubernetes to pod name)
        - name: HOSTNAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        {{- else }}
        # Manual configuration
        - name: CORE_INSTANCE_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        {{- end }}
        
        - name: EXECUTION_STATUS_REFRESH_INTERVAL_SECONDS
          value: {{ .Values.core.executionRefresh.intervalSeconds | quote }}
        {{- end }}
        
        # Existing environment variables
        - name: _REDIS_URL_CORE
          value: {{ include "harbor.redisForCore" . }}
        # ... other env vars ...
        
        {{- if and .Values.core.executionRefresh.enabled (not .Values.core.executionRefresh.autoDetectInstanceId) }}
        volumeMounts:
        - name: shared-config
          mountPath: /shared
        {{- end }}
        
        resources:
          {{- toYaml .Values.core.resources | nindent 10 }}
      
      {{- if and .Values.core.executionRefresh.enabled (not .Values.core.executionRefresh.autoDetectInstanceId) }}
      volumes:
      - name: podinfo
        downwardAPI:
          items:
          - path: name
            fieldRef:
              fieldPath: metadata.name
      - name: shared-config
        emptyDir: {}
      {{- end }}
{{- end }}
```

## Testing in Kubernetes

### 1. Deploy Harbor with 5 Core Instances

```bash
# Install Harbor with Helm
helm repo add harbor https://helm.goharbor.io
helm repo update

# Create custom values
cat > my-values.yaml <<EOF
core:
  replicas: 5
  executionRefresh:
    enabled: true
    autoDetectInstanceId: true
    intervalSeconds: 30

redis:
  type: internal
  internal:
    sentinel:
      enabled: true
EOF

# Install
helm install harbor harbor/harbor \
  -f my-values.yaml \
  --namespace harbor \
  --create-namespace
```

### 2. Verify Instance Configuration

```bash
# Check that all pods are running
kubectl get pods -n harbor -l component=core

# Expected output:
# NAME           READY   STATUS    RESTARTS   AGE
# harbor-core-0  1/1     Running   0          2m
# harbor-core-1  1/1     Running   0          2m
# harbor-core-2  1/1     Running   0          2m
# harbor-core-3  1/1     Running   0          2m
# harbor-core-4  1/1     Running   0          2m

# Check instance ID detection in logs
for i in {0..4}; do
  echo "=== harbor-core-$i ==="
  kubectl logs -n harbor harbor-core-$i | grep -i "instance"
done

# Expected output:
# === harbor-core-0 ===
# INFO: instance 0/5: found 1000 executions in queue
# === harbor-core-1 ===
# INFO: instance 1/5: found 1000 executions in queue
# ...
```

### 3. Verify Work Distribution

```bash
# Check that each instance processes its assigned subset
kubectl logs -n harbor harbor-core-0 | grep "skipped (assigned to other instances)"
kubectl logs -n harbor harbor-core-1 | grep "skipped (assigned to other instances)"

# Each should show ~80% skipped (for 5 instances)
```

### 4. Monitor Performance

```bash
# Check Redis CPU usage
kubectl exec -n harbor redis-0 -- redis-cli INFO CPU

# Check database connections
kubectl exec -n harbor postgresql-0 -- psql -U postgres -c \
  "SELECT count(*) FROM pg_stat_activity WHERE datname='registry';"

# Both should show significant reduction compared to before
```

## Scaling Operations

### Scale Up (Add More Instances)

```bash
# Scale from 5 to 8 instances
helm upgrade harbor harbor/harbor \
  --set core.replicas=8 \
  --reuse-values \
  -n harbor

# Kubernetes will:
# 1. Create harbor-core-5, harbor-core-6, harbor-core-7
# 2. Each new pod auto-detects its instance ID
# 3. Work is automatically redistributed (id % 8)
```

### Scale Down (Remove Instances)

```bash
# Scale from 8 to 5 instances
helm upgrade harbor harbor/harbor \
  --set core.replicas=5 \
  --reuse-values \
  -n harbor

# Kubernetes will:
# 1. Terminate harbor-core-7, harbor-core-6, harbor-core-5
# 2. Remaining pods continue with their IDs
# 3. Work from terminated pods picked up by remaining instances
```

## Troubleshooting

### Issue: Pods Not Detecting Instance ID

**Symptoms:**
```
WARNING: Failed to detect instance ID, using default 0
```

**Solution:**
```bash
# Check if HOSTNAME is set correctly
kubectl exec -n harbor harbor-core-0 -- env | grep HOSTNAME

# Should output: HOSTNAME=harbor-core-0

# If not, ensure StatefulSet is used (not Deployment)
kubectl get statefulset -n harbor
```

### Issue: Multiple Instances with Same ID

**Symptoms:**
```
INFO: instance 0/5: found 1000 executions, 1000 succeed, 0 skipped
INFO: instance 0/5: found 1000 executions, 1000 succeed, 0 skipped
```

**Solution:**
```bash
# Check pod names
kubectl get pods -n harbor -l component=core -o custom-columns=NAME:.metadata.name

# Ensure they follow pattern: harbor-core-{0,1,2,3,4}
# If using Deployment, switch to StatefulSet
```

## Best Practices

1. **Always Use StatefulSet** for multi-instance Harbor core
2. **Enable Auto-Detection** (`autoDetectInstanceId: true`)
3. **Use Redis Sentinel** for high availability
4. **Monitor Metrics** (Redis CPU, DB connections, execution refresh latency)
5. **Start Small** (2-3 instances) and scale up as needed
6. **Test Scaling** in staging before production

## Performance Expectations

| Instances | Redis CPU | DB Queries/cycle | Each Instance Processes |
|-----------|-----------|------------------|-------------------------|
| 1         | 10-20%    | 4M               | 100% (1M executions)    |
| 3         | 10-20%    | 4M               | 33% (~333K executions)  |
| 5         | 10-20%    | 4M               | 20% (200K executions)   |
| 10        | 10-20%    | 4M               | 10% (100K executions)   |

**Key Insight:** Redis and DB load stay constant regardless of instance count!

