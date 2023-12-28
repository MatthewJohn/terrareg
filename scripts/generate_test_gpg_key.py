#!python
import gnupg
import tempfile

with tempfile.TemporaryDirectory() as temp_dir:
    gpg = gnupg.GPG(gnupghome=temp_dir, keyring=None, use_agent=False)
    key = gpg.gen_key(gpg.gen_key_input(key_type="RSA", key_length=1024, passphrase='password'))

    print(f"""
            {{
                # {key.fingerprint}
                "ascii_armor": \"""
{key.gpg.export_keys(key.fingerprint)}
\""".strip(),
            }}
    """)