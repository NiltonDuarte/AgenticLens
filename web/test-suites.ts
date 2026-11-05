// Pure file owners are explicit; all other supported test files are integration.
// A Node annotation alone does not make filesystem/process tests into units.
export const unitTestFiles = [
  'src/domain/**/*.{test,spec}.?(c|m)[jt]s?(x)',
  'src/utils/{debugPrompt,theme}.{test,spec}.?(c|m)[jt]s?(x)',
];
