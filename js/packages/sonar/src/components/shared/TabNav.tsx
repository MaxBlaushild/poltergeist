import { TabGroup, TabList, Tab, TabPanels, TabPanel } from '@headlessui/react';
import React from 'react';

export type TabNavProps = {
  tabs: React.ReactNode[];
  children: React.ReactNode;
}

export type TabItemProps = {
  children: React.ReactNode;
  key: string;
}

export const TabNav = ({ tabs, children }: TabNavProps) => {
  return (
    <TabGroup className="w-full">
      <TabList className="flex gap-4 w-full">
        {tabs.map((tab, i) => (
          <Tab
            key={i}
            className="rounded-full py-1 px-3 text-sm/6 font-semibold text-black focus:outline-none data-[selected]:bg-black/10 data-[hover]:bg-black/5 data-[selected]:data-[hover]:bg-black/10 data-[focus]:outline-1 data-[focus]:outline-black"
        >
          {tab}
        </Tab>
        ))}
      </TabList>
      <TabPanels className="mt-3 w-full">
        {children}
      </TabPanels>
  </TabGroup>
  )
};

export const TabItem = ({ children, key }: TabItemProps) => {
  return <TabPanel key={key} className="rounded-xl bg-black/5 p-3">
    {children}
  </TabPanel>
};
