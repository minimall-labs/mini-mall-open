package com.minimall.openbff;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.context.properties.ConfigurationPropertiesScan;

@SpringBootApplication
@ConfigurationPropertiesScan
public class OpenBffApplication {

    public static void main(String[] args) {
        SpringApplication.run(OpenBffApplication.class, args);
    }
}
