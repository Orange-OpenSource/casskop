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
                "1_concepts/2_features",
                "1_concepts/3_design_principes",
                "1_concepts/4_roadmap",
            ],
            "Setup": [
                "2_setup/1_getting_started",
                "2_setup/2_install_plugin",
                "2_setup/3_multi_casskop",
                {
                    "type" : "category",
                    "label": "Platform Setup",
                    "items"  : [
                        "2_setup/2_platform_setup/1_gke",
                        "2_setup/2_platform_setup/2_minikube",
                    ]
                },
            ],
            "Advanced Configuration": [
                "3_configuration_deployment/0_customizable_install_with_helm",
                "3_configuration_deployment/1_cassandra_cluster",
                "3_configuration_deployment/2_cassandra_configuration",
                "3_configuration_deployment/3_storage",
                "3_configuration_deployment/4_sidecars",
                "3_configuration_deployment/5_kubernets_objects",
                "3_configuration_deployment/6_cpu_memory_usage",
                "3_configuration_deployment/7_cluster_topology",
                "3_configuration_deployment/8_implementation_architecture",
                "3_configuration_deployment/9_advanced_configuration",
                "3_configuration_deployment/10_nodes_management",
                "3_configuration_deployment/11_cassandra_cluster_status",
            ],
            "Examples": [],
            "Operations" : [
                "5_operations/1_cluster_operations",
                "5_operations/2_pods_operations",
                "5_operations/3_multi_casskop",
                "5_operations/4_upgrade_operator",
                "5_operations/5_upgrade_bootstrap_image",
                "5_operations/6_uninstall_casskop",
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