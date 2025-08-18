BEGIN;

INSERT INTO language (language, title, description)
  VALUES ('rego', 'Open Policy Agent / Rego', 'https://www.openpolicyagent.org/docs/policy-language');

INSERT INTO language (language, title, description)
  VALUES ('cedar', 'Cedar', 'https://docs.cedarpolicy.com/');

INSERT INTO language (language, title, description)
  VALUES ('cerbos', 'Cerbos/CEL', 'https://docs.cerbos.dev/cerbos/latest/policies/index.html');

INSERT INTO language (language, title, description)
  VALUES ('openfga', 'OpenFGA', 'https://github.com/openfga/language');

COMMIT;
