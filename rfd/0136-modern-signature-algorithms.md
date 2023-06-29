---
authors: Nic Klaassen (nic@goteleport.com)
state: draft
---

# RFD 136 - Modern Signature Algorithms

## Required Approvers

* Engineering: (@jakule || @espadolini)
* Security: (@reedloden || @jentfoo) && Doyensec
* Product: @klizhentas

## What

Teleport should support modern key types and signature algorithms, currently
only RSA2048 keys are supported with the PKCS#1 v1.5 signature scheme.
This applies to CA keys and subject (user/host/db) keys, but each can/will be
addressed individually.

## Why

Modern algorithms like ECDSA and Ed25519 offer better security properties with
smaller keys that are faster to generate and sign with.
Some of the more restrictive security policies are starting to reject RSA2048
(e.g. [RHEL 8's FUTURE policy](https://access.redhat.com/articles/3642912)).

## Details

### Summary

We will introduce a new config to `teleport.yaml` and `cluster_auth_preference`
to control the key types and signature algorithms used by Teleport CAs and all
clients and hosts which have certificates issued by those CAs.

This config will default to a `recommended` set of algorithms for each protocol
chosen by us to balance security, compatibility, and performance.
We will reserve the right to change this set of `recommended` algorithms when
either:

* the major version of the auth server's teleport.yaml config changes, or
* in a major version release of Teleport.

Most Teleport administrators will never need to see or interact with this config
because they can trust that we will select a vetted set of standards-compliant
algorithms that are trusted to be secure, and we will not break compatibility
with internal Teleport components or third-party software unless deemed
absolutely necessary for security reasons.

Teleport administrators will be able to deviate from the `recommended`
algorithms when they have a compliance need (they must use a particular
algorithm) or a compatibility need (one of our selected algorithms is not
supported by an external softare that interacts with Teleport in their
deployment).

Here is what the config will look like in its default state:

```yaml
ca_key_params:
  user:
    ssh:
      algorithm: recommended
      allowed_subject_algorithms: [recommended]
    tls:
      algorithm: recommended
      allowed_subject_algorithms: [recommended]
  host:
    ssh:
      algorithm: recommended
      allowed_subject_algorithms: [recommended]
    tls:
      algorithm: recommended
      allowed_subject_algorithms: [recommended]
  db:
    tls:
      algorithm: recommended
      allowed_subject_algorithms: [recommended]
  openssh:
    ssh:
      algorithm: recommended
      allowed_subject_algorithms: [recommended]
  jwt:
    jwt:
      algorithm: recommended
      allowed_subject_algorithms: [recommended]
  saml_idp:
    tls:
      algorithm: recommended
      allowed_subject_algorithms: [recommended]
  oidc_idp:
    jwt:
      algorithm: recommended
      allowed_subject_algorithms: [recommended]
```

You can imagine that if we had this config today, the `recommended` keyword here
would expand to `RSA2048_PKCS1_SHA(256|512)` for all protocols.

When we are ready to update the defaults (in a major config version or a major
release) we will update the `recommended` rules to default to the following:

(Note: there will be no actual change to the configuration resource which will
still show `recommended`, the actual values will be computed within Teleport)

(Reviewers: I need your help confirming if these selections are "good" or
recommending alternatives)

```yaml
ca_key_params:
  user:
    ssh:
      algorithm: Ed25519
      # RSA2048 will initially be allowed for older `tsh` clients that don't
      # know how to generate Ed25519 certs, and removed in a future major version
      allowed_subject_algorithms: [Ed25519, RSA2048_PKCS1_SHA512]
    tls:
      algorithm: ECDSA_P256_SHA256
      # RSA2048 will initially be allowed for older `tsh` clients that don't
      # know how to generate Ed25519 certs, and removed in a future major version
      allowed_subject_algorithms: [ECDSA_P256_SHA256, RSA2048_PKCS1_SHA256]
  host:
    ssh:
      algorithm: Ed25519
      # RSA2048 will initially be allowed for older hosts that don't know how to
      # generate Ed25519 certs, and removed in a future major version
      allowed_subject_algorithms: [Ed25519, RSA2048_PKCS1_SHA512]
    tls:
      algorithm: ECDSA_P256_SHA256
      # RSA2048 will initially be allowed for older hosts that don't know how to
      # generate Ed25519 certs, and removed in a future major version
      allowed_subject_algorithms: [ECDSA_P256_SHA256, RSA2048_PKCS1_SHA256]
  db:
    tls:
      # multiple DBs only support RSA so it will remain the default for now
      algorithm: RSA3072_PKCS1_SHA256
      # db certs are often fairly long-lived so we should prefer a larger key
      # size for them.
      # We will allow Ed25519 for connections the Proxy makes to Teleport
      # database services because they are short lived, generated often, and
      # only used internally within Teleport components.
      allowed_subject_algorithms: [RSA3072_PKCS1_SHA256, RSA2048_PKCS1_SHA256, Ed25519]
  openssh:
    ssh:
      algorithm: Ed25519
      # RSA2048 will initially be allowed for older hosts that don't know how to
      # generate Ed25519 certs, and removed in a future major version
      allowed_subject_algorithms: [Ed25519, RSA2048_PKCS1_SHA512]
  jwt:
    jwt:
      algorithm: ECDSA_P256_SHA256
      allowed_subject_algorithms: [ECDSA_P256_SHA256]
  saml_idp:
    tls:
      algorithm: ECDSA_P256_SHA256
      allowed_subject_algorithms: [ECDSA_P256_SHA256]
  oidc_idp:
    jwt:
      algorithm: ECDSA_P256_SHA256
      allowed_subject_algorithms: [ECDSA_P256_SHA256]
```

For backward-compatibility, all certs already signed by trusted CAs will
continue to be trusted, `allowed_subject_algorithms` can be modified at any time
without breaking connectivity, and only controls the allowed algorithms used for
new certificates signed by the CA.

Changing CA `algorithm` values in this config will take effect for:

* new Teleport clusters
* existing Teleport clusters only after a CA rotation.

### Algorithms

These algorithms are being considered for support:

#### RSA

Private key sizes: 2048, 3072, 4096

Signature algorithms: PKCS#1 v1.5

Digest/hash algorithms: SHA512 for SSH, SHA256 for TLS

Considerations:

* RSA2048 is the current default and deviating from it by default may break
  compatibility with third-party components and protocols
* RSA has the most widespread support among all protocols
* Certain database protocols only support RSA client certs
  * <https://docs.snowflake.com/en/user-guide/key-pair-auth#step-2-generate-a-public-key>
* Some apps may only support RSA signed JWTs
* If we must continue to support RSA, we might as well support larger key sizes
  (at least for CA keys), 3072 and 4096-bit are the most commonly used and
  supported by e.g. GCP KMS.
* golang.org/x/crypto/ssh uses SHA512 hash by default with all RSA public keys
  (but this can be overridden)
  <https://github.com/golang/crypto/blob/0ff60057bbafb685e9f9a97af5261f484f8283d1/ssh/certs.go#L443-L445>
* crypto/x509 uses SHA256 hash by default with all RSA public keys
  (but this can be overridden)
  <https://github.com/golang/go/blob/dbf9bf2c39116f1330002ebba8f8870b96645d87/src/crypto/x509/x509.go#L1411-L1414>
* ssh only supports the PKCS#1 v1.5 signature scheme with RSA keys
  <https://datatracker.ietf.org/doc/html/rfc8332>
* FIPS 186-5 approves all listed options
* BoringCrypto supports all listed options
* We could consider PSS signatures instead of PKCS#1 v1.5 for TLS and JWT
  signatures, but SSH does not support it.

#### ECDSA

Curves: P-256

Digest/hash algorithms: SHA256

Considerations:

* ECDSA has good support across SSH and TLS protocols for both client and CA
  certs.
* ECDSA certs are supported by web browsers.
* ECDSA key generation is *much* faster than RSA key generation.
* ECDSA signatures are faster than RSA signatures.
* FIPS 186-5 approves all listed options
* BoringCrypto supports all listed options
* The P-256 curve is the most popular, it is considered to be secure, and it has
  the broadest support among external tools.
* We could consider supporting the P-384 and P-521 curves for CAs.

#### EdDSA

Curves: Ed25519

Digest/hash algorithms: none (the full message is signed without hashing)

Considerations:

* There is widespread support for Ed25519 SSH certs.
* Go libraries support Ed25519 for TLS
* Support for Ed25519 is *not* widespread in the TLS ecosystem.
* YubiHSM and GCP KMS do *not* support Ed25519 keys.
* Ed25519 is considered by some to be the fastest, most secure, most modern
  option for SSH certs.
* Ed25519 key generation is *much* faster than RSA key generation.
* Ed25519 signatures are faster than RSA signatures.
* FIPS 186-5 approves Ed25519
* BoringCrypto does not support Ed25519
  <https://go.googlesource.com/go/+/dev.boringcrypto/src/crypto/tls/boring.go#80>
* Ed25519 is the only EdDSA curve supported in the Go standard library.

#### Summary

* We are probably forced to continue unconditionally using RSA for database
  certs, I'm assuming this would apply to both client and CA.
* Ed25519 is a modern favourite for SSH, but TLS (and HSM, KMS) support is lacking.
* Teleport CAs use separate keypairs for SSH and TLS, they do not need to use
  the same algorithm.
* Teleport derives client SSH and TLS certs from the same client keypair,
  supporting different algorithms for each will require larger changes.
* It seems it is time to split client SSH and TLS keys to support the popular
  and secure Ed25519 algorithm for SSH and the widely-suported
  `ECDSA_P256_SHA256` algorithm for TLS. This will also allows to evolve the
  algorithms used for each protocol independently in the future.

### CAs

Each Teleport CA holds 1 or more of the following:

* SSH public and private key
* TLS certificate and private key
* JWT public and private key

Each CA key may be a software key stored in the Teleport backend, an HSM key
held in an HSM connected to a local Auth server via a PKCS#11 interface, or a
KMS key held in GCP KMS.
In the future we will likely support more KMS services.

Teleport currently has these CAs:

#### User CA

keys: ssh, tls

uses: user ssh cert signing, user tls cert signing, ssh hosts trust this CA

* current SSH algo: `RSA2048_PKCS1_SHA512`
* proposed supported SSH algos:
  * `Ed25519`
  * `ECDSA_P256_SHA256`
  * `RSA2048_PKCS1_SHA512`
  * `RSA3072_PKCS1_SHA512`
  * `RSA4096_PKCS1_SHA512`
* proposed supported SSH `allowed_subject_algorithms`:
  * `Ed25519`
  * `ECDSA_P256_SHA256`
  * `RSA2048_PKCS1_SHA512`
  * `RSA3072_PKCS1_SHA512`
  * `RSA4096_PKCS1_SHA512`
* reasoning:
  * `Ed25519` is the current best-in-class for SSH
  * `ECDSA_P256_SHA256` has Go BoringCrypto support
  * some environments still require RSA

* current TLS algo: `RSA2048_PKCS1_SHA256`
* proposed supported TLS algos:
  * `ECDSA_P256_SHA256`
  * `Ed25519`
  * `RSA2048_PKCS1_SHA256`
* proposed supported TLS `allowed_subject_algorithms`:
  * `ECDSA_P256_SHA256`
  * `Ed25519`
  * `RSA2048_PKCS1_SHA512`
  * `RSA3072_PKCS1_SHA512`
  * `RSA4096_PKCS1_SHA512`
* reasoning:
  * `ECDSA_P256_SHA256` has the broadest support among external tools
  * `Ed25519` support is becoming more common and some prefer it
  * `RSA2048_PKCS1_SHA256` will be supported for backward compatibility, but
    ECDSA support is widespread enough I don't think we should support bigger
    RSA keys, we should guide people toward ECDSA instead.

#### Host CA

keys: ssh, tls

uses: host ssh cert signing, host tls cert signing, ssh clients trust this CA

* current SSH algo: `RSA2048_PKCS1_SHA512`
* proposed supported SSH algos:
  * `Ed25519`
  * `ECDSA_P256_SHA256`
  * `RSA2048_PKCS1_SHA512`
  * `RSA3072_PKCS1_SHA512`
  * `RSA4096_PKCS1_SHA512`
* proposed supported SSH `allowed_subject_algorithms`:
  * `Ed25519`
  * `ECDSA_P256_SHA256`
  * `RSA2048_PKCS1_SHA512`
  * `RSA3072_PKCS1_SHA512`
  * `RSA4096_PKCS1_SHA512`
* reasoning:
  * `Ed25519` is the current best-in-class for SSH
  * `ECDSA_P256_SHA256` has Go BoringCrypto support
  * some environments still require RSA

* current TLS algo: `RSA2048_PKCS1_SHA256`
* proposed supported TLS algos:
  * `ECDSA_P256_SHA256`
  * `Ed25519`
  * `RSA2048_PKCS1_SHA256`
* proposed supported TLS `allowed_subject_algorithms`:
  * `ECDSA_P256_SHA256`
  * `Ed25519`
  * `RSA2048_PKCS1_SHA256`
* reasoning:
  * `EDCSA_P256_SHA256` has BoringCrypto support.
  * `Ed25519` support is becoming more common and some people prefer it
  * `RSA2048_PKCS1_SHA256` will be supported for backward compatibility

#### Database CA

keys: tls

uses:

* signs (often) long-lived db cert used to authenticate db to database service
* signs short-lived proxy cert used to authenticate proxy to database service
* signed snowflake JWTs

* current TLS algo: `RSA2048_PKCS1_SHA256`
* proposed supported TLS algos:
  * `ECDSA_P256_SHA256`
  * `Ed25519`
  * `RSA2048_PKCS1_SHA256`
  * `RSA3072_PKCS1_SHA256`
  * `RSA4096_PKCS1_SHA256`
* proposed supported TLS `allowed_subject_algorithms`:
  * `ECDSA_P256_SHA256`
  * `Ed25519`
  * `RSA2048_PKCS1_SHA256`
  * `RSA3072_PKCS1_SHA256`
  * `RSA4096_PKCS1_SHA256`
* reasoning:
  * some database protocols still require RSA, reduce friction by keeping it as
    the default
  * eventually we should default to `RSA3072_PKCS1_SHA256` for long-lived db
    certs that require RSA, but use an algorithm that is cheaper to generate
    keys for the proxy certs that authenticate to the db service, since these
    are only used internally we should go for Ed25519

#### OpenSSH Host CA

keys: ssh

uses: signs user certs to authenticate to registered OpenSSH nodes, registered
OpenSSH nodes trust this CA.

* current SSH algo: `RSA2048_PKCS1_SHA512`
* proposed supported SSH algos:
  * `Ed25519`
  * `ECDSA_P256_SHA256`
  * `RSA2048_PKCS1_SHA512`
  * `RSA3072_PKCS1_SHA512`
  * `RSA4096_PKCS1_SHA512`
* proposed supported SSH `allowed_subject_algorithms`:
  * `Ed25519`
  * `ECDSA_P256_SHA256`
  * `RSA2048_PKCS1_SHA512`
* reasoning:
  * `Ed25519` is the current best-in-class for SSH
  * `ECDSA_P256_SHA256` has Go BoringCrypto support
  * some environments still require RSA

#### JWT CA

keys: jwt

uses: user jwt cert signing, exposed at `/.well-known/jwks.json`, applications
that verify user JWTs trust this CA

* current algo: `RSA2048_PKCS1_SHA256`
* proposed supported algos:
  * `ECDSA_P256_SHA256`
  * `Ed25519`
  * `RSA2048_PKCS1_SHA256`
* proposed supported TLS `allowed_subject_algorithms`:
  * `ECDSA_P256_SHA256`
  * `Ed25519`
  * `RSA2048_PKCS1_SHA256`
* reasoning:
  * except RSA, `EDCSA_P256_SHA256` has the broadest support among external tools
  * `Ed25519` support is becoming more common and some people prefer it
  * `RSA2048_PKCS1_SHA256` will be supported for backward compatibility, but
    ECDSA support is widespread enough I don't think we should support bigger
    RSA keys, we should guide people toward ECDSA instead.

#### OIDC IdP CA

keys: jwt

uses: signing JWTs as an OIDC provider.

* current algo: `RSA2048_PKCS1_SHA256`
* proposed supported algos:
  * `ECDSA_P256_SHA256`
  * `Ed25519`
  * `RSA2048_PKCS1_SHA256`
* proposed supported TLS `allowed_subject_algorithms`:
  * `ECDSA_P256_SHA256`
  * `Ed25519`
  * `RSA2048_PKCS1_SHA256`
* reasoning:
  * except RSA, `EDCSA_P256_SHA256` has the broadest support among external tools
  * `Ed25519` support is becoming more common and some people prefer it
  * `RSA2048_PKCS1_SHA256` will be supported for backward compatibility, but
    ECDSA support is widespread enough I don't think we should support bigger
    RSA keys, we should guide people toward ECDSA instead.

#### SAML IdP CA

keys: tls

uses: signing SAML assertions as a SAML provider.

* current TLS algo: `RSA2048_PKCS1_SHA256`
* proposed supported TLS algos:
  * `ECDSA_P256_SHA256`
  * `Ed25519`
  * `RSA2048_PKCS1_SHA256`
* proposed supported TLS `allowed_subject_algorithms`:
  * `ECDSA_P256_SHA256`
  * `Ed25519`
  * `RSA2048_PKCS1_SHA256`
* reasoning:
  * `EDCSA_P256_SHA256` has BoringCrypto support.
  * `Ed25519` support is becoming more common and some people prefer it
  * `RSA2048_PKCS1_SHA256` will be supported for backward compatibility

### CA Configuration

CA key types and signature algorithms will be configurable via
`cluster_auth_preference` and `teleport.yaml`

We want it configurable via `teleport.yaml` so that you can start a new cluster
and the CA keys will be automatically generated at first start with the correct
algorithms, so you don't have to immediately edit the `cap` and then rotate all
of your brand-new CAs.

We want it configurable via `cluster_auth_preference` as well so that it can be
configurable for Cloud users.

If any values under `ca_key_params` are explicitly set in the
`cluster_auth_preference`, it will completely override the settings from
`teleport.yaml`

```yaml
# teleport.yaml
version: v3
auth_service:
  enabled: true

  ca_key_params:

    # ca_key_params is already a part of the `teleport.yaml`, it has `gcp_kms` and
    # `pkcs11` subsections for enabling KMS/HSMs per auth server.
    gcp_kms:
      keyring: projects/teleport-dev-123456/locations/us-west1/keyRings/nic-example-1
      protection_level: "SOFTWARE"

    user:
      ssh:
        # any supported algorithm can be selected for each protocol per CA
        algorithm: Ed25519

        # any subset of supported subject algorithms can be allowed, up to date
        # compliant clients should select the first algorithm from this list that
        # they support, or a preffered algorithm for the protocol from this list
        allowed_subject_algorithms:
          - Ed25519
          - RSA2048_PKCS1_SHA512
      tls:
        # use recommended (the default) to automatically select Teleport's
        # recommended algorithm for this CA and protocol
        algorithm: recommended
        allowed_subject_algorithms:
          # this will expand to our recommended list of allowed subject
          # algorithms
          - recommended
    host:
      ssh:
        # this configures the host CA to use an RSA4096 key
        algorithm: RSA4096_PKCS1_SHA512
        # this configures hosts to always get RSA2048 certs
        allowed_subject_algorithms:
          - RSA2048_PKCS1_SHA512

      # any unlisted fields will default to all "recommended"
      tls:
        algorithm: recommended
        allowed_subject_algorithms: [recommended]

    # any unlisted CAs will default to all "recommended"
    db:
      tls:
        algorithm: recommended
        allowed_subject_algorithms: [recommended]
    openssh:
      ssh:
        algorithm: recommended
        allowed_subject_algorithms: [recommended]
    jwt:
      jwt:
        algorithm: recommended
        allowed_subject_algorithms: [recommended]
    saml_idp:
      tls:
        algorithm: recommended
        allowed_subject_algorithms: [recommended]
    oidc_idp:
      jwt:
        algorithm: recommended
        allowed_subject_algorithms: [recommended]
```

```yaml
kind: cluster_auth_preference
metadata:
  name: cluster-auth-preference
spec:
  ca_key_params:
    user:
      ssh:
        algorithm: recommended
        allowed_subject_algorithms: [recommended]
      tls:
        algorithm: recommended
        allowed_subject_algorithms: [recommended]
    host:
      ssh:
        algorithm: recommended
        allowed_subject_algorithms: [recommended]
      tls:
        algorithm: recommended
        allowed_subject_algorithms: [recommended]
    db:
      tls:
        algorithm: recommended
        allowed_subject_algorithms: [recommended]
    openssh:
      ssh:
        algorithm: recommended
        allowed_subject_algorithms: [recommended]
    jwt:
      jwt:
        algorithm: recommended
        allowed_subject_algorithms: [recommended]
    saml_idp:
      tls:
        algorithm: recommended
        allowed_subject_algorithms: [recommended]
    oidc_idp:
      jwt:
        algorithm: recommended
        allowed_subject_algorithms: [recommended]
```

#### Cloud

Cloud will be able to select their preferred defaults by configuring them in the
`teleport.yaml`.
Cloud users will be able to change the CA algorithms by modifying the
`cluster_auth_preference` and performing a CA rotation.

### Subjects (users/hosts/DBs/JWTs)

It will be great to update the CA key algorithms for security and performance
benefits, but what users really see on a day-to-day basis is their user keys.

"Subjects" that have certificates issued by the Teleport CAs include:

* Teleport users via `tsh login`
* Teleport users via `tsh app login`
* Teleport users via `tsh db login`
* Teleport users via Teleport Connect
* Teleport services (ssh, app, db, kube, windows desktop, etc)
* Machine ID (`tbot`)
* Teleport Plugins
* OpenSSH hosts
* Databases
* Proxies communicating with remote database services

All of these currently generate an RSA2048 keypair locally, send the public key
to the auth server, and receive signed certificates of some variety.

The `allowed_subject_algorithms` field described in the CA configuration section
will control which key algorithms can be used by subject keys.
These should be in preference order, compliant clients should pick the first
algorithm from the list that they support.
This will allow us to introduce new algorithms to the list without breaking
compatibility with existing clients.

In some cases, Teleport services may choose a preferred algorithm from the list
that is not the first one they support.
For example, the Database CA signs DB certs that are often fairly long-lived, as
well as short-lived certs used by proxies connecting to remote database services
on behalf of the user, it may be preferable to select different algorithms for
each case.

### Splitting user SSH and TLS private keys

In order to support the use of different key types for SSH and TLS, users will
need to start generating two different private keys, one for each protocol.
This will change the disk layout of the ~/.tsh directory:

```diff
  ~/.tsh/                             --> default base directory
  ├── current-profile                 --> file containing the name of the currently active profile
  ├── one.example.com.yaml            --> file containing profile details for proxy "one.example.com"
  ├── two.example.com.yaml            --> file containing profile details for proxy "two.example.com"
  ├── known_hosts                     --> trusted certificate authorities (their keys) in a format similar to known_hosts
  └── keys                            --> session keys directory
     ├── one.example.com              --> Proxy hostname
     │   ├── certs.pem                --> TLS CA certs for the Teleport CA
-    │   ├── foo                      --> Private Key for user "foo"
-    │   ├── foo.pub                  --> Public Key
+    │   ├── foo                      --> SSH Private Key for user "foo"
+    │   ├── foo.pub                  --> SSH Public Key
     │   ├── foo.ppk                  --> PuTTY PPK-formatted keypair for user "foo"
     │   ├── kube_credentials.lock    --> Kube credential lockfile, used to prevent excessive relogin attempts
+    │   ├── foo-x509-privkey.pem     --> TLS client private key
     │   ├── foo-x509.pem             --> TLS client certificate for Auth Server
     │   ├── foo-ssh                  --> SSH certs for user "foo"
     │   │   ├── root-cert.pub        --> SSH cert for Teleport cluster "root"
     │   │   └── leaf-cert.pub        --> SSH cert for Teleport cluster "leaf"
     │   ├── foo-app                  --> App access certs for user "foo"
     │   │   ├── root                 --> App access certs for cluster "root"
     │   │   │   ├── appA-x509.pem    --> TLS cert for app service "appA"
     │   │   │   └── appB-x509.pem    --> TLS cert for app service "appB"
     │   │   │   └── appB-localca.pem --> Self-signed localhost CA cert for app service "appB"
     │   │   └── leaf                 --> App access certs for cluster "leaf"
     │   │       └── appC-x509.pem    --> TLS cert for app service "appC"
     │   ├── foo-db                   --> Database access certs for user "foo"
     │   │   ├── root                 --> Database access certs for cluster "root"
     │   │   │   ├── dbA-x509.pem     --> TLS cert for database service "dbA"
     │   │   │   ├── dbB-x509.pem     --> TLS cert for database service "dbB"
     │   │   │   └── dbC-wallet       --> Oracle Client wallet Configuration directory.
     │   │   ├── leaf                 --> Database access certs for cluster "leaf"
     │   │   │   └── dbC-x509.pem     --> TLS cert for database service "dbC"
     │   │   └── proxy-localca.pem    --> Self-signed TLS Routing local proxy CA
     │   ├── foo-kube                 --> Kubernetes certs for user "foo"
     │   |    ├── root                 --> Kubernetes certs for Teleport cluster "root"
     │   |    │   ├── kubeA-kubeconfig --> standalone kubeconfig for Kubernetes cluster "kubeA"
     │   |    │   ├── kubeA-x509.pem   --> TLS cert for Kubernetes cluster "kubeA"
     │   |    │   ├── kubeB-kubeconfig --> standalone kubeconfig for Kubernetes cluster "kubeB"
     │   |    │   ├── kubeB-x509.pem   --> TLS cert for Kubernetes cluster "kubeB"
     │   |    │   └── localca.pem      --> Self-signed localhost CA cert for Teleport cluster "root"
     │   |    └── leaf                 --> Kubernetes certs for Teleport cluster "leaf"
     │   |        ├── kubeC-kubeconfig --> standalone kubeconfig for Kubernetes cluster "kubeC"
     │   |        └── kubeC-x509.pem   --> TLS cert for Kubernetes cluster "kubeC"
     |   └── cas                       --> Trusted clusters certificates
     |        ├── root.pem             --> TLS CA for teleport cluster "root"
     |        ├── leaf1.pem            --> TLS CA for teleport cluster "leaf1"
     |        └── leaf2.pem            --> TLS CA for teleport cluster "leaf2"
     └── two.example.com               --> Additional proxy host entries follow the same format
                ...
```

RPCs such as `GenerateUserCerts` will also need to change to support passing
both of the public keys along.
These will remain backward compatible by continuing to use the single public key
for both protocols if both are not passed.

### HSMs/KMS

We will attempt to use the configured CA algorithms when the CA keys are backed
by HSMs or KMS services.
If the specific algorithm is not supported, we will do our best to return an
informative error message to the user.
It will be the reponsibility of the Teleport admin to select an algorithm
supported by their particular HSM/KMS.

### Backward Compatibility

* Will the change impact older clients? (tsh, tctl)

By default, Auth servers with non-default algorithms configured should continue
to sign certificates for clients on older Teleport versions using RSA2048 keys.
This can be configured with a list of `allowed_subject_algorithms` per CA key
type so that RSA keys can eventually be rejected when the user is ready.

* Are there any backend migrations required?

A CA rotation will effectively act as the backend migration when changing
algorithms.

### Remote Clusters

TODO: figure out compatibility for remote clusters using different algorithms or
running different Teleport versions.

### Security

This entire RFD is about improving Teleport's security.

We are not introducing any new endpoints, CLI commands, or APIs.

All supported algorithms will be reviewed by our internal security team and
external security auditors.

We will only use Go standard library implementations of crypto algorithms (or
BoringCrypto if compiled in FIPS mode).

### UX

New configuration in `teleport.yaml` and `cluster_auth_preference` are described
above.

### Proto Specification

TODO:

Include any `.proto` changes or additions that are necessary for your design.

### Audit Events

User login and existing certificate generation events will be supplemented with
info on the algorithms used by the subject and the CA.

### Observability

Log messages will be emitted whenever CA keys/certs are generated or rotated,
including the algorithms used for all new keys.

Audit events will also include the algorithms used.

### Product Usage

We will not add telemetry or usage events, logs and audit events will indicate
if this feature is being used.

### Test Plan

TODO:

Include any changes or additions that will need to be made to
the [Test Plan](../.github/ISSUE_TEMPLATE/testplan.md) to appropriately
test the changes in your design doc and prevent any regressions from
happening in the future.
