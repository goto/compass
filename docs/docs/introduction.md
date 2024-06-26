---
id: introduction
slug: /
---

# Introduction

Welcome to the introductory guide to Compass! This guide is the best place to start with Compass. We cover what Compass is, what problems it can solve, how it works, and how you can get started using it. If you are familiar with the basics of Compass, the guide provides a more detailed reference of available features.

## What is Compass?

Compass is a search and discovery engine built for querying application deployments, datasets and meta resources. It can also optionally track data flow relationships between these resources and allow the user to view a representation of the data flow graph.

![](/assets/overview.svg)

## The problem we aim to solve

Organizational teams face the challenge of finding the right data from various sources for their analysis and decision-making needs. A lack of transparency about the flow of data within the organization can lead to problems when it comes to updating or discarding outdated data.

Manual methods for tracking the dependencies of data are time-consuming and subject to human error or oversight as it depends on mapping of the movement of data in the organisation on knowledge in people's head. Huge code volume, rate of change and complexity in manually examning the data changes make the process unsustainable.
In addition, organizations may struggle to locate the most relevant data from the massive amounts stored in their data stores. The longer it takes for users to find the business data they need, the less productive they are.

To address these challenges, Compass was designed as a data discovery and search tool for organizations. It provides comprehensive asset-listing and search capabilities to enhance user productivity. Organizing the data assets using Compass allows the data professionals to collect, access, and enrich metadata to support data discovery and governance. The data lineage information provided by Compass also helps organizations meet compliance requirements by offering a clear record of the movement of sensitive data. Using Compass one can identify and star the most relevant data assets, and safely delete when you don’t need.

## Key Features

Discover why users choose Compass as their main data discovery and lineage service

- **Full text search** Faster and better search results powered by ElasticSearch full text search capability.
- **Search Tuning** Narrow down your search results by adding filters, getting your crisp results.
- **Data Lineage** Understand the relationship between metadata with data lineage interface.
- **Scale:** Compass scales in an instant, both vertically and horizontally for high performance.
- **Extensibility:** Add your own metadata types and resources to support wide variety of metadata.
- **Runtime:** Compass can run inside VMs or containers in a fully managed runtime environment like kubernetes.

## Using Compass
You can manage the data discovery for your resources via Compass in any of the following ways:
### Compass Command Line Interface

You can use the Compass command line interface to issue commands and to perform the entire data discovery flow. Using the command line can be faster and more convenient than the console. For more information on using the Compass CLI, see the [CLI Reference](./reference/cli.md) page.

### HTTPS API

You can get hands on asset searching, listing, tagging, starring and much more by using the Compass HTTPS API, which lets you issue HTTPS requests directly to the service. When you use the HTTPS API, you must include the user information from the configurable identity uuid in the request header.
For more information, see the [User Management](./concepts/user.md) and [API reference](./reference/api.md) pages.

## Where to go from here

See the [installation](./installation) page to install the Compass CLI. Next, we recommend completing the guides. The tour provides an overview of most of the existing functionality of Compass and takes approximately 20 minutes to complete.

After completing the tour, check out the remainder of the documentation in the reference and concepts sections for your specific areas of interest. We've aimed to provide as much documentation as we can for the various components of Compass to give you a full understanding of Compass's surface area.

Finally, follow the project on [GitHub](https://github.com/goto/compass), and contact us if you'd like to get involved.