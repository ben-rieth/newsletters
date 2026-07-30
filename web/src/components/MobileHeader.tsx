import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import type { ReactNode } from 'react';
import { createPortal } from 'react-dom';
import type { LinkProps } from '@tanstack/react-router';

export type MobileHeaderConfig = {
  title?: string;
  back?: LinkProps;
};

type MobileHeaderStore = {
  config: MobileHeaderConfig;
  claim: (owner: object, config: MobileHeaderConfig | null) => void;
  actionSlot: HTMLElement | null;
  registerActionSlot: (el: HTMLElement | null) => void;
};

const MobileHeaderContext = createContext<MobileHeaderStore | null>(null);

const useStore = () => {
  const store = useContext(MobileHeaderContext);
  if (!store) {
    throw new Error('MobileHeader components require MobileHeaderProvider');
  }
  return store;
};

export const MobileHeaderProvider = ({ children }: { children: ReactNode }) => {
  const [entry, setEntry] = useState<{
    owner: object;
    config: MobileHeaderConfig;
  } | null>(null);
  const [actionSlot, setActionSlot] = useState<HTMLElement | null>(null);

  const claim = useCallback(
    (owner: object, config: MobileHeaderConfig | null) => {
      setEntry((current) => {
        if (config === null) {
          return current?.owner === owner ? null : current;
        }
        return { owner, config };
      });
    },
    [],
  );

  const value = useMemo(
    () => ({
      config: entry?.config ?? {},
      claim,
      actionSlot,
      registerActionSlot: setActionSlot,
    }),
    [entry, claim, actionSlot],
  );

  return (
    <MobileHeaderContext.Provider value={value}>
      {children}
    </MobileHeaderContext.Provider>
  );
};

export const useMobileHeaderStore = useStore;

export const useMobileHeader = (config: MobileHeaderConfig) => {
  const { claim } = useStore();
  const owner = useRef({}).current;
  const serialized = JSON.stringify(config);

  useEffect(() => {
    claim(owner, JSON.parse(serialized) as MobileHeaderConfig);
    return () => claim(owner, null);
  }, [claim, owner, serialized]);
};

export const MobileHeaderAction = ({ children }: { children: ReactNode }) => {
  const { actionSlot } = useStore();
  if (!actionSlot) return null;
  return createPortal(children, actionSlot);
};
