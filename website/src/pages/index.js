import React from 'react';
import classnames from 'classnames';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import useBaseUrl from '@docusaurus/useBaseUrl';
import styles from './styles.module.css';


const features = [
    {
        title: <>Kubernetes Operator</>,
        imageUrl: 'img/operator-sdk.png',
        description: (
            <>
                CassKop will define a new Kubernetes object named CassandraCluster which will be used to describe
                and instantiate a Cassandra Cluster in Kubernetes.
            </>
        ),
    },
    {
        title: <>Open-Source</>,
        imageUrl: 'img/open_source.svg',
        description: (
            <>
                Open source software released under the Apache 2.0 license.
            </>
        ),
    },
    {
        title: <>Cassandra Cluster in K8S</>,
        imageUrl: 'img/kubernetes.png',
        description: (
            <>
                CassKop is a Kubernetes custom controller which will provide you with the required Kubernetes resources for your deployment.
            </>
        ),
    },
    {
        title: <>Space Scoped</>,
        imageUrl: 'img/namespace.png',
        description: (
            <>
                CassKop is listening only in the Kubernetes namespace it is deployed in, and
                is able to manage several Cassandra Clusters within this namespace.
            </>
        ),
    },
    {
        title: <>Operate Cassandra Node with Kubectl plugin!</>,
        imageUrl: 'img/cassandra.png',
        description: (
            <>
                Casskop manage a list of operations at 2 levels :
                Cluster operations with CRD and pod operations with our kubectl plugin!
                Pod operation example: cleanup, rebuild...
            </>
        ),
    },
    {
        title: <>Multi-Datacenter Deployment</>,
        imageUrl: 'img/dc.png',
        description: (
            <>
                Multi-CassKop (included) can deploy one Cassandra cluster on different Kubernetes clusters deployed on different regions for real HA.
            </>
        ),
    }
];

function Feature({imageUrl, title, description}) {
    const imgUrl = useBaseUrl(imageUrl);
    return (
        <div className={classnames('col col--4', styles.feature)}>
            {imgUrl && (
                <div className="text--center">
                    <img className={styles.featureImage} src={imgUrl} alt={title} />
                </div>
            )}
            <h3>{title}</h3>
            <p>{description}</p>
        </div>
    );
}

function Home() {
    const context = useDocusaurusContext();
    const {siteConfig: {customFields = {}} = {}} = context;

    return (
        <Layout permalink="/" description={customFields.description}>
            <div className={styles.hero}>
                <div className={styles.heroInner}>
                    <h1 className={styles.heroProjectTagline}>
                        <img
                            alt="Casskop"
                            className={styles.heroLogo}
                            src={useBaseUrl('img/casskop_alone.png')}
                        />
                        Open-Source, Apache <span className={styles.heroProjectKeywords}>Cassandra</span>{' '}
                        operator for <span className={styles.heroProjectKeywords}>Kubernetes</span>{' '}
                    </h1>
                    <div className={styles.indexCtas}>
                        <Link
                            className={styles.indexCtasGetStartedButton}
                            to={useBaseUrl('docs/2_setup/1_getting_started')}>
                            Get Started
                        </Link>
                        <span className={styles.indexCtasGitHubButtonWrapper}>
              <iframe
                  className={styles.indexCtasGitHubButton}
                  src="https://ghbtns.com/github-btn.html?user=Orange-OpenSource&amp;repo=casskop&amp;type=star&amp;count=true&amp;size=large"
                  width={160}
                  height={30}
                  title="GitHub Stars"
              />
            </span>
                    </div>
                </div>
            </div>
            <div className={classnames(styles.announcement, styles.announcementDark)}>
                <div className={styles.announcementInner}>
                    The <span className={styles.heroProjectKeywords}>CassKop</span> Cassandra Kubernetes operator makes it <span className={styles.heroProjectKeywords}>easy</span> to run Apache Cassandra on Kubernetes.
                    Apache Cassandra is a popular, free, open-source, distributed wide column store, <span className={styles.heroProjectKeywords}>NoSQL database</span> management system.
                    The operator allows to <span className={styles.heroProjectKeywords}>easily create and manage racks and data centers</span> aware Cassandra clusters.
                </div>
            </div>
            <div className={styles.section}>
                {features && features.length && (
                    <section className={styles.features}>
                        <div className="container">
                            <div className="row">
                                {features.map((props, idx) => (
                                    <Feature key={idx} {...props} />
                                ))}
                            </div>
                        </div>
                    </section>
                )}
            </div>
        </Layout>
    );
}

export default Home;
