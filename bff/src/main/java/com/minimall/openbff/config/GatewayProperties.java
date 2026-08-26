package com.minimall.openbff.config;

import org.springframework.boot.context.properties.ConfigurationProperties;

@ConfigurationProperties(prefix = "minimall.gateway")
public record GatewayProperties(String baseUrl) {
}
