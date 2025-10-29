# Namecheap Webhook Configuration Guide

This guide explains how to configure and use Namecheap webhooks with the provider-namecheap Crossplane provider.

## Overview

The provider-namecheap supports receiving webhook events from Namecheap for real-time notifications about domain, DNS, SSL, and account changes. This enables immediate reconciliation and status updates without polling.

## Supported Webhook Events

### Domain Events
- `domain.registered` - Domain successfully registered
- `domain.renewed` - Domain renewal completed
- `domain.expired` - Domain has expired
- `domain.transferred` - Domain transfer completed

### DNS Events
- `dns.record.created` - DNS record created
- `dns.record.updated` - DNS record modified
- `dns.record.deleted` - DNS record removed

### SSL Events
- `ssl.issued` - SSL certificate issued
- `ssl.renewed` - SSL certificate renewed
- `ssl.expired` - SSL certificate expired
- `ssl.revoked` - SSL certificate revoked

### Account Events
- `account.updated` - Account information changed
- `payment.received` - Payment processed successfully
- `payment.failed` - Payment processing failed

## Provider Configuration

### 1. Enable Webhook Support

Configure the provider deployment with webhook settings:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: provider-namecheap
  namespace: crossplane-system
spec:
  template:
    spec:
      containers:
      - name: provider
        image: ghcr.io/rossigee/provider-namecheap:v0.5.3
        env:
        # Enable webhook server
        - name: WEBHOOK_ENABLED
          value: "true"
        - name: WEBHOOK_PORT
          value: "8443"
        - name: WEBHOOK_PATH
          value: "/webhook"

        # Webhook security
        - name: WEBHOOK_SECRET
          valueFrom:
            secretKeyRef:
              name: namecheap-webhook-secret
              key: webhook-secret

        # TLS configuration
        - name: WEBHOOK_TLS_CERT_DIR
          value: "/tmp/k8s-webhook-server/serving-certs"
        - name: TLS_SERVER_CERTS_DIR
          value: "/tmp/k8s-webhook-server/serving-certs"

        ports:
        - name: webhook
          containerPort: 8443
          protocol: TCP
        - name: metrics
          containerPort: 8080
          protocol: TCP

        volumeMounts:
        - name: webhook-certs
          mountPath: /tmp/k8s-webhook-server/serving-certs
          readOnly: true

      volumes:
      - name: webhook-certs
        secret:
          secretName: namecheap-webhook-certs
```

### 2. Create Webhook Secret

Generate a secure secret for webhook signature verification:

```bash
# Generate a random webhook secret
WEBHOOK_SECRET=$(openssl rand -hex 32)

# Create the secret in Kubernetes
kubectl create secret generic namecheap-webhook-secret \
  --from-literal=webhook-secret="$WEBHOOK_SECRET" \
  -n crossplane-system
```

### 3. Configure TLS Certificates

Generate TLS certificates for the webhook server:

```bash
# Generate TLS certificates
openssl req -x509 -newkey rsa:4096 -keyout tls.key -out tls.crt -days 365 -nodes \
  -subj "/CN=provider-namecheap-webhook.crossplane-system.svc.cluster.local"

# Create TLS secret
kubectl create secret tls namecheap-webhook-certs \
  --cert=tls.crt \
  --key=tls.key \
  -n crossplane-system
```

### 4. Expose Webhook Service

Create a service to expose the webhook endpoint:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: provider-namecheap-webhook
  namespace: crossplane-system
spec:
  selector:
    app: provider-namecheap
  ports:
  - name: webhook
    port: 443
    targetPort: 8443
    protocol: TCP
  type: LoadBalancer  # or NodePort/ClusterIP with Ingress
```

### 5. Configure Ingress (Optional)

If using an ingress controller:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: namecheap-webhook
  namespace: crossplane-system
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/backend-protocol: HTTPS
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - namecheap-webhook.example.com
    secretName: namecheap-webhook-tls
  rules:
  - host: namecheap-webhook.example.com
    http:
      paths:
      - path: /webhook
        pathType: Prefix
        backend:
          service:
            name: provider-namecheap-webhook
            port:
              number: 443
```

## Namecheap Portal Configuration

### 1. Access Webhook Settings

1. Log into your Namecheap account
2. Navigate to **Domain List** > **Manage**
3. Go to **Advanced DNS** > **Webhooks** (if available)
4. Or contact Namecheap support to enable webhook functionality

### 2. Configure Webhook Endpoint

Add your webhook endpoint URL:

```
URL: https://namecheap-webhook.example.com/webhook
Method: POST
Content-Type: application/json
```

### 3. Set Webhook Secret

Configure the same secret you created in Kubernetes:

```
Secret: <your-webhook-secret>
Signature Header: X-Namecheap-Signature
Signature Method: HMAC-SHA256
```

### 4. Select Events

Choose which events to receive:

- ✅ Domain registration events
- ✅ Domain renewal events
- ✅ Domain expiration events
- ✅ DNS record changes
- ✅ SSL certificate events
- ✅ Account updates
- ✅ Payment notifications

### 5. Test Webhook

Use the test function in Namecheap portal to verify connectivity:

```json
{
  "id": "test-12345",
  "type": "domain.registered",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "domain": "example.com",
    "registrant": "user@example.com"
  }
}
```

## Monitoring and Troubleshooting

### Health Checks

Monitor webhook server health:

```bash
# Check webhook server health
curl -k https://namecheap-webhook.example.com/health

# Expected response:
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "processors": [
    "domain.registered",
    "dns.record.created",
    "ssl.issued"
  ]
}
```

### Metrics

Access webhook metrics:

```bash
# Get webhook metrics
curl -k https://namecheap-webhook.example.com/metrics

# Expected response:
{
  "requests_total": 42,
  "requests_errors": 0,
  "events_processed": 38,
  "processing_errors": 0,
  "request_duration_avg": 0.150,
  "request_count": 42,
  "uptime_seconds": 3600
}
```

### Log Analysis

Check provider logs for webhook events:

```bash
# View webhook logs
kubectl logs -n crossplane-system deployment/provider-namecheap | grep webhook

# Look for successful events:
# INFO webhook event received {"event_id": "123", "event_type": "domain.registered"}
# INFO successfully processed webhook event {"event_id": "123"}

# Look for errors:
# ERROR invalid webhook signature
# ERROR failed to process webhook event
```

### Common Issues

#### 1. Signature Verification Failures

```bash
# Check webhook secret matches
kubectl get secret namecheap-webhook-secret -n crossplane-system -o jsonpath='{.data.webhook-secret}' | base64 -d
```

#### 2. TLS Certificate Issues

```bash
# Verify TLS certificate
kubectl get secret namecheap-webhook-certs -n crossplane-system -o yaml

# Test TLS connectivity
openssl s_client -connect namecheap-webhook.example.com:443 -servername namecheap-webhook.example.com
```

#### 3. Network Connectivity

```bash
# Test from within cluster
kubectl run test-pod --image=curlimages/curl --rm -it -- \
  curl -k https://provider-namecheap-webhook.crossplane-system.svc.cluster.local/health
```

#### 4. Webhook Not Receiving Events

1. Verify webhook URL is accessible from internet
2. Check Namecheap portal webhook configuration
3. Ensure events are enabled for your account type
4. Test webhook endpoint manually:

```bash
# Test webhook endpoint
curl -X POST https://namecheap-webhook.example.com/webhook \
  -H "Content-Type: application/json" \
  -H "X-Namecheap-Signature: sha256=<calculated-signature>" \
  -d '{"id":"test","type":"domain.registered","timestamp":"2024-01-01T12:00:00Z","data":{"domain":"test.com"}}'
```

## Security Considerations

### 1. Network Security

- Use TLS for all webhook traffic
- Implement network policies to restrict access
- Use valid SSL certificates (Let's Encrypt recommended)

### 2. Authentication

- Always verify webhook signatures
- Use strong, randomly generated secrets
- Rotate webhook secrets regularly

### 3. Authorization

- Limit webhook access to necessary IP ranges
- Monitor for suspicious activity
- Implement rate limiting at the ingress level

### 4. Data Protection

- Log webhook events appropriately (avoid sensitive data)
- Implement audit trails for webhook processing
- Ensure compliance with data protection regulations

## Advanced Configuration

### Custom Event Processors

Create custom processors for specific business logic:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: webhook-processors
  namespace: crossplane-system
data:
  custom-processor.yaml: |
    processors:
      domain.registered:
        - type: email-notification
          recipients: ["admin@example.com"]
        - type: slack-notification
          webhook: "https://hooks.slack.com/..."
        - type: metrics-update
          endpoint: "https://metrics.example.com/webhook"
```

### High Availability

Configure multiple webhook endpoints:

```yaml
# In Namecheap portal, configure multiple webhooks:
# Primary: https://webhook-primary.example.com/webhook
# Secondary: https://webhook-secondary.example.com/webhook
```

### Load Balancing

Use multiple provider replicas:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: provider-namecheap
spec:
  replicas: 3  # Multiple replicas for HA
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
```

This comprehensive setup ensures reliable webhook processing with proper security, monitoring, and high availability.