[comment]: <> (TODO: paste cool logo here)

## What is Tiny Apps

Tiny Apps is a platform for easily testing & deploying dashboards & web applications (ex. Streamlit) to Kubernetes.
Implemented as a Kubernetes CRD (Custom Resource Definition).
Currently supports Streamlit and Dash - with Gradio support coming soon.

## Highlights
- Test & deploy Streamlit & Dash apps to Kubernetes with ease.
- API endpoints for managing Tiny App lifecycle.
- Integrated with Prometheus to track & serve app metrics.
- Check out [jupyterlab-tinyapp]("link") - JupyterLab extension that allows users to test & deploy
their notebooks as Tiny Apps.

## Getting Started
See [Getting Started](docs/getting_started.md).

## Contributing
Please review [CONTRIBUTING.md](CONTRIBUTING.md) for how to contribute to this project.

## Roadmap
- Integrate with LDAP & OAuth for authentication & authorization to manage/access app.
- UI for managing apps.
- Scale apps based on usage.
- VSCode extension for testing & deploying apps.
- Support for more web frameworks.
