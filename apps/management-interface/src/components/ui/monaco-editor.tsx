import {Editor, Monaco, OnMount} from '@monaco-editor/react';
import {editor} from 'monaco-editor';
import {useRef} from 'react';
import {regoLanguageConfig, regoLanguageTokens} from '@/lib/monaco/rego-language';
import {cedarLanguageConfig, cedarLanguageTokens} from '@/lib/monaco/cedar-language';

interface MonacoEditorProps {
  value?: string,
  language?: string,
  readOnly?: boolean,
  height?: string | number,
  className?: string,
  onChange?: (value: string | undefined, ev: editor.IModelContentChangedEvent) => void
}

export function MonacoEditor({
                               value = '',
                               language = 'plaintext',
                               readOnly = true,
                               height = '100%',
                               className,
                               onChange
                             }: MonacoEditorProps) {
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null);
  const languagesRegistered = useRef(false);

  const handleEditorDidMount: OnMount = (editor, monaco: Monaco) => {
    editorRef.current = editor;
    // Register custom languages once
    if (!languagesRegistered.current) {
      // Register Rego
      monaco.languages.register({id: 'rego'});
      monaco.languages.setMonarchTokensProvider('rego', regoLanguageTokens);
      monaco.languages.setLanguageConfiguration('rego', regoLanguageConfig);

      // Register Cedar
      monaco.languages.register({id: 'cedar'});
      monaco.languages.setMonarchTokensProvider('cedar', cedarLanguageTokens);
      monaco.languages.setLanguageConfiguration('cedar', cedarLanguageConfig);

      languagesRegistered.current = true;
    }
  };

  // Normalize language name to lowercase
  const normalizedLanguage = language?.toLowerCase() || 'plaintext';

  return (
    <div className={className}>
      <Editor
        height={height}
        defaultLanguage={normalizedLanguage}
        language={normalizedLanguage}
        value={value}
        theme="vs"
        onChange={onChange}
        options={{
          readOnly,
          lineNumbers: 'on',
          minimap: {enabled: false},
          wordWrap: 'on',
          scrollBeyondLastLine: false,
          automaticLayout: true,
          fontSize: 14,
          padding: {top: 16, bottom: 16},
          scrollbar: {
            vertical: 'visible',
            horizontal: 'visible',
            useShadows: true,
            verticalScrollbarSize: 10,
            horizontalScrollbarSize: 10,
          },
        }}
        onMount={handleEditorDidMount}
      />
    </div>
  );
}
