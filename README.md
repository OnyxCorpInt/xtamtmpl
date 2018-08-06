# xtamtmpl

XTAM Template, or `xtamtmpl`, is a shell script that parameterizes configuration files (called templates) with records from [Xton Access Manager (XTAM)](https://www.xtontech.com/).

## Usage

Configuration values may be supplied to `xtamtmpl` by a combination (in order of precedence) of flags, environment variables and configuration files. These values can be listed with the `--help` flag:

	$ ./xtamtmpl --help
	Usage of ./xtamtmpl:
	  -config="": path to config file
	  -output-path="/etc/config": Directory to which filled templates will be written
	  -template-path="/mnt/templates": Directory from which to read templates
	  -xtam-cas-host="": XTAM CAS URL (required)
	  -xtam-container-id="": XTAM container (folder or Vault) ID (required)
	  -xtam-host="": XTAM Base URL (required)
	  -xtam-password="": XTAM authentication string (required)
	  -xtam-username="": XTAM authentication string (required)

	  
Some defaults are provided that may be useful for containerized environments.

## Configuration

Flag | Usage
--- | ---
Template Path | Directory from which templates will be read. Only files ending in `.template` will be used, and the extension will be dropped to determine the output filename. e.g. `templates/cts.json.template` will become `output/cts.json`. Note that this is not recursive: a flat directory structure is expected.
Output Path | Directory to which the final configuration files will be written. Any files in this directly whose names clash with written configuration will be overwritten, while non-clashing files will be left alone. So it should be safe to "merge" templated config in with other files.
XTAM Host | The base URL for the XTAM server. **Do** include the protocol and path XTAM is deployed to, e.g. `https://my.domain.name/xtam`, but omit elements of the path specific to the REST API, e.g. `/rest`.
XTAM Container ID | The ID of the folder or Vault from which secrets and certificates will be read. The authenticating user must be permitted to read this container.
CAS Host | The base URL for the CAS authentication server. Do not include the trailing `/login`.
XTAM Username | A CAS user authorized to use the provided XTAM server. A service ticket will be obtained from CAS for this user on each invocation.
XTAM Password | Password of the authenticating user.

### Example CLI Invocation

Configuration from multiple sources can be combined as per the rules laid out by the [namsral/flag package](https://github.com/namsral/flag):

	$ cat sample.config
	template-path /mnt/templates
	output-path /etc/config/myapp
	
	$ XTAM_USERNAME=doug XTAM_PASSWORD=redacted ./xtamtmpl -xtam-container-id=180 -config sample.config

The precedence for flags (lowest to highest) are: default values, config file, environment variables, then CLI flags.

### Example Kubernetes Invocation

The primary use case of this utility is to be embedded in a Docker image, which then runs as an init container in a Kubernetes pod. In other words, we want to pull secrets from XTAM, and stick them into an app's configuration files before it starts.

We're doing this because Kubernetes doesn't provide a config templating mechanism and the default protections offered by the `Secret` type are lacking.

Below is a mostly complete manifest for an example container using `xtamtmpl` to template its config. Afterward, we'll break it down one piece at a time and see how all they all come together.

```
kind: Secret
apiVersion: v1
metadata:
  name: xtamtmpl-config
data:
  configfile: eHRhbS11c2VybmFtZSBkb3VnICAgICAgICAgICAgICAgICAgICAgIAp4dGFtLXBhc3N3b3JkIHJlZGFjdGVkCnh0YW0tY2FzLWhvc3QgaHR0cHM6Ly9teW9yZy5jb20vY2FzCnh0YW0taG9zdCBodHRwczovL215b3JnLmNvbS94dGFtCnh0YW0tZm9sZGVyLWlkIDEyMzQ1Cg==
---
apiVersion: apps/v1
kind: Pod
metadata:
  name: example
spec:
  volumes:
    - name: template-volume
      configMap:
        name: example-templates
    - name: xtamtmpl-config-volume
      secret:
        secretName: xtamtmpl-config
    - name: config-volume
      emptyDir: {}
  initContainers:
  - name: init-config
    image: xtamtmpl:latest
    command: ['xtamtmpl', '-config', '/etc/xtamtmpl/configfile', '-output-path', '/etc/config/']
    volumeMounts:
    - name: template-volume
      mountPath: /mnt/templates
    - name: xtamtmpl-config-volume
      mountPath: /etc/xtamtmpl
    - name: config-volume
      mountPath: /etc/config
  containers:
  - name: example
    image: some-image
    args: ["/etc/config/example.json"]
    volumeMounts:
    - name: config-volume
      mountPath: /etc/config

```

Okay, so the minimal example isn't very, uh, minimal. Let's break it down:

```
kind: Secret
apiVersion: v1
metadata:
  name: xtamtmpl-config
data:
  configfile: <snipped base64 encoded key value pairs>
```

It's helpful to put all of the XTAM configuration that will be common to all init containers into a shared Kubernetes `Secret` object. This could include credentials and XTAM server URLs. Depending on how your XTAM records are organized, you can also include the container ID. An example of the xtamtmpl config file:

---

```
xtam-username doug
xtam-password redacted
xtam-cas-host https://myorg.com/cas
xtam-host https://myorg.com/xtam
xtam-container-id 12345
```

Which would then be base64-encoded and used as the value of `configfile`. There are [other ways to create Secret objects](https://kubernetes.io/docs/concepts/configuration/secret/#creating-your-own-secrets), so take your pick.

---

```
volumes:
  - name: config-volume
    emptyDir: {}
  - name: template-volume
    configMap:
      name: example-templates
  - name: xtamtmpl-config-volume
    secret:
      secretName: xtamtmpl-config
```

Several volumes must come together in the init container:

* `config-volume` - we use a simple `emptyDir` for the location of the final processed configuration files. This could also be a volume containing other files that don't need processing.
* `template-volume` - this is a `ConfigMap` containing any templates `xtamtmpl` should process. These files should have whatever final name is desired in the target directory, plus a `.template` suffix. So `example.json.template` from the template volume will become `example.json` in `config-volume`
* `xtamtmpl-config-volume` - since the `xtamtmpl` utility itself has a configuration file, we bring it in from a `Secret` object here.

---

```
...
spec:
  initContainers:
  - name: init-config
    image: xtamtmpl:latest
    command: ['xtamtmpl', '-config', '/etc/xtamtmpl/configfile', '-output-path', '/etc/config/']
```

The pod spec includes an `initContainers` section, which is how `xtamtmpl` swoops in to generate the app's config from templates before the app container will start. We pass in a few parameters specific to this pod:

* `-config` the path to the config file, which is where we will mount `xtamtmpl-config-volume` plus the name of the common config file (see ` xtamtmpl-config`)
* `-output-path` - the application-specific directory where the final config files will be read from. Note that `xtamtmpl` is indisciminate in overwriting files, so it's best to point this at an `emptyDir`.

---

```
volumeMounts:
- name: template-volume
  mountPath: /mnt/templates
- name: xtamtmpl-config-volume
  mountPath: /etc/xtamtmpl
- name: config-volume
  mountPath: /etc/config
```

The volume mounts mostly follow from the above:

* `template-volume` - by default, `xtamtmpl` reads all templates in the `/mnt/templates` directory.
* `xtamtmpl-config-volume` - The `xtamtmpl` common config file should be mounted where it will be found by the init container command.
* `config-volume` this mount is shared by the init container and the app container, and is where `xtamtmpl` will write the final configuration files

---

```
containers:
- name: example
  image: some-image
  args: ["/etc/config/example.json"]
  volumeMounts:
  - name: config-volume
    mountPath: /etc/config
```

At last comes your application container. The details will depend on what you're trying to run, but note that all of your final configuration files (e.g. `example.json` above) are guaranteed to be written before your container is created, otherwise the init container will fail and halt container creation.

#### Troubleshooting

`xtamtmpl` will try to bail early on any error (XTAM auth, invalid templates, missing directories) with a helpful message. One exception to this is if the `xtamtmpl` configuration file (`-config`) is missing or invalid, the app aborts with exit code 2 and no message. 

As the init container itself is fleeting, troubleshooting becomes a little more difficult.

Logs from the last failure of `xtamtmpl` can be obtained with the handy `-c` flag of `kubectl logs`:

```
$ kubectl logs pod/example -c init-config
flag provided but not defined: -oops
Usage of xtamtmpl:
...
```

Sometimes you may want to inspect the contents of the init container, for example, to see if all your templates are mounted and available. One easy way to do this is to instruct the container to sleep instead of invoking `xtamtmpl`:

```
  initContainers:
  - name: init-config
    image: xtamtmpl:latest
    command: ['sleep', '600']
```
Then you can invoke a shell inside the "sleeping" init container and poke around:

```
$ kubectl exec -it example -c init-config -- sh
# cd /mnt/templates
# ls
example.json.template
# _
```



## Templating

`xtamtmpl` uses the ~~hideous~~ simple [text/template](https://golang.org/pkg/text/template/#hdr-Actions) package for templating. Rather than repeat what is covered there, here's a list of the functions exposed to templates (invoked with dot syntax):

* `Secret(string)`: Fetches a Secret record by name and writes it to the template. Note that because `xtamtmpl` is format-agnostic, no escaping of the secret value is performed. So take care that secret values won't mangle whatever format your config file uses. For example, secrets going into a JSON string value shouldn't contain double quotes. 

* `CertPEM(string)`: Fetches a Certificate record by name and writes its PEM-encoded representation to the template. This is mostly useful when the template consists solely of a single call to `CertPEM`, but PEM's flexible nature could allow multiple certifcates to be concatencated or a single cert to be appended to an existing chain, etc.

### Example Templates

To see how this comes together, let's consider a template embedded into a Kubernetes `ConfigMap`:

```
kind: ConfigMap
apiVersion: v1
metadata:
  name: example-templates
data:
  example.json.template: |-
    {
    	"AccessToken": "{{.Secret "example-access-token"}}",
    	...
    }
  example.pem.template: |-
    {{.CertPEM "example-cerificate"}}
```

Here we see the files that `xtamtmpl` will process as they are suffixed by `.template`:

* `example.json.template` has a JSON field `AccessToken` whose value comes from that of a Secret record with name `example-access-token`. It will be fetched and inserted into the template, between the double-quotes of the value.

* `example.pem.template` consists of nothing but a call to `CertPEM`, which injects the certificate verbatim. So it is essentially like the certificate named `example-cerificate` is copied to `example.pem` without modification.

## Building

Run `make` in the project root directory to produce an `xtamtmpl` executable.

Run `make image` to build a local Docker image, tagged with the version found in `./VERSION`.

## Limitations

* Right now all secrets and certificates to read must come from a single XTAM container per invocation of the tool.
* Authentication can only be performed with [CAS](https://apereo.github.io/cas/5.0.x/installation/Service-Management.html), and the authenticating user must have permission to read the container.
* All templates are read from a single directory, and written to a single target directory.
* No escaping is done (or possible) of values coming from XTAM and inserted into templates.