package com.minimall.openbff.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.client.RestClient;

@Configuration
public class GatewayClientConfig {

    @Bean
    RestClient gatewayRestClient(GatewayProperties gatewayProperties) {
        return RestClient.builder()
                .baseUrl(gatewayProperties.baseUrl())
                .build();
    }
}
