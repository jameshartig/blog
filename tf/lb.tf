resource "google_compute_backend_service" "run" {
  project = var.project_id
  name    = "blog"

  load_balancing_scheme = "EXTERNAL_MANAGED"
  protocol              = "HTTP"
  enable_cdn            = false
  compression_mode      = "DISABLED"
  session_affinity      = "CLIENT_IP"
  edge_security_policy  = null
  security_policy       = null

  dynamic "backend" {
    for_each = google_compute_region_network_endpoint_group.blog
    content {
      group           = backend.value.self_link
      capacity_scaler = 1.0
    }
  }

  log_config {
    enable        = true
    sample_rate   = "1.0"
    optional_mode = "INCLUDE_ALL_OPTIONAL"
  }

  iap {
    enabled              = false
    oauth2_client_id     = null
    oauth2_client_secret = null
  }
}

resource "google_compute_url_map" "redirect" {
  project = var.project_id
  name    = "redirect"

  default_url_redirect {
    https_redirect         = true
    redirect_response_code = "MOVED_PERMANENTLY_DEFAULT"
    strip_query            = false
  }
}

resource "google_compute_url_map" "blog" {
  project = var.project_id

  name            = "blog"
  default_service = google_compute_backend_service.run.self_link
}

resource "google_compute_global_address" "blog_ipv4" {
  project    = var.project_id
  name       = "blog-ipv4"
  ip_version = "IPV4"

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_compute_global_address" "blog_ipv6" {
  project    = var.project_id
  name       = "blog-ipv6"
  ip_version = "IPV6"

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_compute_target_http_proxy" "redirect" {
  project = var.project_id
  name    = "http-redirect"
  url_map = google_compute_url_map.redirect.self_link
}

resource "google_certificate_manager_dns_authorization" "jameshartig_dev" {
  project = var.project_id
  name    = "jameshartig-dev"
  domain  = "jameshartig.dev"

  depends_on = [module.enabled_google_apis]

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_certificate_manager_certificate" "jameshartig_dev" {
  project = var.project_id
  name    = "jameshartig-dev"
  scope   = "DEFAULT"

  managed {
    domains = [
      google_certificate_manager_dns_authorization.jameshartig_dev.domain
    ]
    dns_authorizations = [
      google_certificate_manager_dns_authorization.jameshartig_dev.id
    ]
  }

  lifecycle {
    ignore_changes = [
      labels,
      managed[0].authorization_attempt_info[0].state
    ]
  }
}

resource "google_certificate_manager_dns_authorization" "jameshartig_com" {
  project = var.project_id
  name    = "jameshartig-com"
  domain  = "jameshartig.com"

  depends_on = [module.enabled_google_apis]

  lifecycle {
    prevent_destroy = true
  }
}

resource "google_certificate_manager_certificate" "jameshartig_com" {
  project = var.project_id
  name    = "jameshartig-com"
  scope   = "DEFAULT"

  managed {
    domains = [
      google_certificate_manager_dns_authorization.jameshartig_com.domain
    ]
    dns_authorizations = [
      google_certificate_manager_dns_authorization.jameshartig_com.id
    ]
  }

  lifecycle {
    ignore_changes = [
      labels,
      managed[0].authorization_attempt_info[0].state
    ]
  }
}

# TODO: need to use a certificate map with external LB until https://github.com/hashicorp/terraform-provider-google/issues/17176
resource "google_certificate_manager_certificate_map" "blog" {
  project = var.project_id
  name    = "blog"
}

resource "google_certificate_manager_certificate_map_entry" "jameshartig_dev" {
  project = var.project_id
  name    = "jameshartig-dev"
  map     = google_certificate_manager_certificate_map.blog.name
  certificates = [
    google_certificate_manager_certificate.jameshartig_dev.id
  ]
  hostname = "jameshartig.dev"
}

resource "google_certificate_manager_certificate_map_entry" "jameshartig_com" {
  project = var.project_id
  name    = "jameshartig-com"
  map     = google_certificate_manager_certificate_map.blog.name
  certificates = [
    google_certificate_manager_certificate.jameshartig_com.id
  ]
  hostname = "jameshartig.com"
}

resource "google_compute_ssl_policy" "tls_1_2_modern" {
  project         = var.project_id
  name            = "tls-1-2-modern-policy"
  min_tls_version = "TLS_1_2"
  profile         = "MODERN"
}

resource "google_compute_target_https_proxy" "blog" {
  project        = var.project_id
  name           = "blog-https"
  url_map        = google_compute_url_map.blog.self_link
  ssl_policy     = google_compute_ssl_policy.tls_1_2_modern.self_link
  tls_early_data = "STRICT"

  #certificate_manager_certificates = [
  #  "//certificatemanager.googleapis.com/${google_certificate_manager_certificate.jameshartig_dev.id}"
  #]
  certificate_map = "//certificatemanager.googleapis.com/${google_certificate_manager_certificate_map.blog.id}"
}

resource "google_compute_global_forwarding_rule" "blog_ipv4_http" {
  project               = var.project_id
  name                  = "blog-ipv4-http"
  target                = google_compute_target_http_proxy.redirect.self_link
  ip_address            = google_compute_global_address.blog_ipv4.address
  port_range            = "80"
  load_balancing_scheme = "EXTERNAL_MANAGED"
}

resource "google_compute_global_forwarding_rule" "blog_ipv4_https" {
  project               = var.project_id
  name                  = "blog-ipv4-https"
  target                = google_compute_target_https_proxy.blog.self_link
  ip_address            = google_compute_global_address.blog_ipv4.address
  port_range            = "443"
  load_balancing_scheme = "EXTERNAL_MANAGED"
}

resource "google_compute_global_forwarding_rule" "blog_ipv6_http" {
  project               = var.project_id
  name                  = "blog-ipv6-http"
  target                = google_compute_target_http_proxy.redirect.self_link
  ip_address            = google_compute_global_address.blog_ipv6.address
  port_range            = "80"
  load_balancing_scheme = "EXTERNAL_MANAGED"
}

resource "google_compute_global_forwarding_rule" "blog_ipv6_https" {
  project               = var.project_id
  name                  = "blog-ipv6-https"
  target                = google_compute_target_https_proxy.blog.self_link
  ip_address            = google_compute_global_address.blog_ipv6.address
  port_range            = "443"
  load_balancing_scheme = "EXTERNAL_MANAGED"
}
