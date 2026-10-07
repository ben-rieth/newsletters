import { useRef } from 'react';
import { Button } from '#/components/ui/button';
import useImportNewsletters from '../queries/hooks/useImportNewsletters';

export const ImportNewslettersButton = () => {
  const inputRef = useRef<HTMLInputElement>(null);
  const importNewsletters = useImportNewsletters();

  return (
    <>
      <input
        ref={inputRef}
        type="file"
        accept="application/json,.json"
        className="hidden"
        aria-hidden="true"
        tabIndex={-1}
        onChange={(event) => {
          const file = event.target.files?.[0];
          event.target.value = '';
          if (file) {
            importNewsletters.mutate(file);
          }
        }}
      />
      <Button
        variant="outline"
        onClick={() => inputRef.current?.click()}
        disabled={importNewsletters.isPending}
      >
        {importNewsletters.isPending ? 'Importing…' : 'Import from file'}
      </Button>
    </>
  );
};
