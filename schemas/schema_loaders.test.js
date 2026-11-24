'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { test } = require('node:test');
const { schemaValidation, yamlReader, resolved } = require('yaml-dsl-spec');

const SCHEMA_FILE_NAME = /^go-publisher-v(\d+)\.(\d+)\.schema\.json$/;

/** The minor schemas in this folder, oldest first. */
function minorSchemaFiles() {
  return fs.readdirSync(__dirname)
    .map((name) => ({ name, version: SCHEMA_FILE_NAME.exec(name) }))
    .filter(({ version }) => version !== null)
    .sort((left, right) => (left.version[1] - right.version[1]) || (left.version[2] - right.version[2]))
    .map(({ name }) => name);
}

function loadedSchema(name) {
  return new schemaValidation.SchemaCache().load(name, fs.readFileSync(path.join(__dirname, name), 'utf8'));
}

/** The problems the schema finds in a configuration's text. */
function problemsIn(schema, file, text) {
  const { node } = yamlReader.parseYaml(file, text, []);
  return schema.problems(resolved.Value.fromNode(node), new schemaValidation.ValidationRules());
}

test('every minor schema loads in the reader editors validate with', () => {
  const names = minorSchemaFiles();
  assert.ok(names.length > 0, 'no schema files found');
  for (const name of names) {
    assert.doesNotThrow(() => loadedSchema(name), name);
  }
});

test("the repository's own configuration passes the newest minor", () => {
  const newest = minorSchemaFiles().at(-1);
  const configuration = path.join(__dirname, '..', '.gopublisher.yml');
  const problems = problemsIn(loadedSchema(newest), '.gopublisher.yml', fs.readFileSync(configuration, 'utf8'));
  assert.deepEqual(problems.map((problem) => problem.diagnostic.message), []);
});

test('a field the schema does not define is a problem', () => {
  const newest = minorSchemaFiles().at(-1);
  const problems = problemsIn(loadedSchema(newest), 'mock.yml', 'version: 2\nmock_field: true\n');
  assert.equal(problems.length, 1);
  assert.match(problems[0].diagnostic.message, /mock_field/);
});
