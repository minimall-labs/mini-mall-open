package com.minimall.openbff.web;

import com.minimall.openbff.config.GatewayProperties;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.client.RestClient;

import java.util.LinkedHashMap;
import java.util.Map;

@RestController
@RequestMapping("/bff/open")
public class HealthController {

    private final GatewayProperties gatewayProperties;

    public HealthController(RestClient gatewayRestClient, GatewayProperties gatewayProperties) {
        this.gatewayProperties = gatewayProperties;
    }

    @GetMapping("/health")
    public Map<String, Object> health() {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("service", "open-bff");
        body.put("status", "UP");
        body.put("gateway", gatewayProperties.baseUrl());
        body.put("note", "Shell project — add ISV auth and scoped APIs here");
        return body;
    }

    @GetMapping("/ping")
    public Map<String, Object> ping() {
        Map<String, Object> body = new LinkedHashMap<>();
        body.put("message", "open platform pong");
        body.put("gateway", gatewayProperties.baseUrl());
        return body;
    }
}
