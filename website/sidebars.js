/**
 * Copyright (c) 2017-present, Facebook, Inc.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

module.exports = {
    "docs":
        {
            "Concepts": [
                "1_concepts/1_introduction",
                "1_concepts/2_design_principes",
                "1_concepts/3_features",
                "1_concepts/4_roadmap",
            ],
            "Setup": [
                "2_setup/1_getting_started",
                {
                    "type" : "category",
                    "label": "Platform Setup",
                    "items"  : [
                        "2_setup/2_platform_setup/1_gke",
                        "2_setup/2_platform_setup/2_minikube",
                    ]
                },
                {
                    "type" : "category",
                    "label": "Install",
                    "items"  : [
                        "2_setup/3_install/1_customizable_install_with_helm",
                        "2_setup/3_install/2_install_plugin",
                    ]
                }
            ],
            "Tasks": [
                "3_tasks/1_operator_description",
                {
                    "type" : "category",
                    "label": "Configuration Deployment",
                    "items"  : [
                        "3_tasks/2_configuration_deployment/1_cassandra_cluster",
                        "3_tasks/2_configuration_deployment/2_cassandra_configuration",
                        "3_tasks/2_configuration_deployment/3_storage",
                        "3_tasks/2_configuration_deployment/4_sidecars",
                        "3_tasks/2_configuration_deployment/5_kubernets_objects",
                        "3_tasks/2_configuration_deployment/6_cpu_memory_usage",
                        "3_tasks/2_configuration_deployment/7_cluster_topology",
                        "3_tasks/2_configuration_deployment/8_implementation_architecture",
                        "3_tasks/2_configuration_deployment/9_advanced_configuration",
                        "3_tasks/2_configuration_deployment/10_nodes_management",
                        "3_tasks/2_configuration_deployment/11_cassandra_cluster_status",
                    ]
                },
            ],
            "Examples": [],
            "Operations" : [
                "5_operations/1_cluster_operations",
                "5_operations/2_pods_operations",
                {
                    "type" : "category",
                    "label": "Upgrade",
                    "items"  : [
                        "5_operations/3_upgrading/1_upgrade_operator",
                        "5_operations/3_upgrading/2_upgrade_bootstrap_image",
                    ]
                },
                {
                    "type" : "category",
                    "label": "Uninstall",
                    "items"  : [
                        "5_operations/4_uninstall/1_casskop",
                    ]
                },
                {
                    "type" : "category",
                    "label": "MultiCasskop",
                    "items"  : [
                        "5_operations/5_multicasskop/1_cassandra_cluster",
                    ]
                },

            ],
            "Reference": [
                {
                    "type" : "category",
                    "label": "Cassandra Cluster",
                    "items"  : [
                        "6_references/1_cassandra_cluster/1_cassandra_cluster",
                        "6_references/1_cassandra_cluster/2_topology",
                        "6_references/1_cassandra_cluster/3_cassandra_cluster_status",
                    ]
                },
                {
                    "type" : "category",
                    "label": "MultiCasskop",
                    "items"  : [
                        "6_references/2_multicasskop/1_multicasskop",
                    ]
                },
            ],
            "Troubleshooting" : [
                "7_troubleshooting/1_operations_issues",
                "7_troubleshooting/2_gke_issues",
            ],
            "Contributing" : [
                "8_contributing/1_developer_guide",
                "8_contributing/2_release_guide",
                "8_contributing/3_reporting_bugs",
                "8_contributing/4_credits",

            ]
        }
};