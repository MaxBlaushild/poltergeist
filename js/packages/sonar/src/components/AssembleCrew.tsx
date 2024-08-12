import './AssembleCrew.css';
import React, { useEffect, useState } from 'react';
import { useAPI } from '@poltergeist/contexts';
import {
  Survey,
  Submission,
  SubmissionAnswer,
  User,
  Activity,
} from '@poltergeist/types';
import {
  Combobox,
  ComboboxButton,
  ComboboxInput,
  ComboboxOption,
  ComboboxOptions,
  Transition,
  Field,
  Label,
  TabGroup,
  TabList,
  Tab,
  TabPanels,
  TabPanel,
  Description,
} from '@headlessui/react';
import { CheckIcon, ChevronDownIcon } from '@heroicons/react/20/solid';
import clsx from 'clsx';
import { Modal, ModalSize } from './shared/Modal.tsx';
import { Chip, ChipType } from './shared/Chip.tsx';
import { Button } from './shared/Button.tsx';
import { useActivityContext } from '../contexts/ActivityContext.tsx';
import PersonListItem from './shared/PersonListItem.tsx';

type SelectedAnswer = {
  id: string;
  phoneNumber: string;
  name: string;
};

export const AssembleCrew: React.FC = () => {
  const { apiClient } = useAPI();
  const { categories } = useActivityContext();
  const [submissions, setSubmissions] = useState<Submission[]>([]);
  const [nameFilters, setNameFilters] = useState<string[]>([]);
  const [activityFilters, setActivityFilters] = useState<string[]>([]);
  const [tempNameFilter, setTempNameFilter] = useState<string>('');
  const [tempActivityFilter, setTempActivityFilter] = useState<string>('');
  const [activityComboboxQuery, setActivityComboboxQuery] = useState('');
  const [nameComboboxQuery, setNameComboboxQuery] = useState('');
  const [notSelectedUsers, setNotSelectedUsers] = useState<User[]>([]);
  const [showCrewModal, setShowCrewModal] = useState(false);

  useEffect(() => {
    const fetchSurveysAndAnswers = async () => {
      try {
        const submissions = await apiClient.get<Submission[]>(
          '/sonar/surveys/submissions'
        );
        setSubmissions(submissions);
      } catch (error) {
        console.error('Failed to fetch surveys and answers', error);
      }
    };

    fetchSurveysAndAnswers();
  }, [apiClient]);

  const handleAddNameFilter = (user): void => {
    if (user) {
      const userName = user.name;
      if (userName && !nameFilters.includes(userName)) {
        setNameFilters([...nameFilters, userName]);
        setTempNameFilter('');
      }
    }
  };

  const handleAddActivityFilter = (activity): void => {
    if (activity) {
      const activityTitle = activity.activity.title;
      if (activityTitle && !activityFilters.includes(activityTitle)) {
        setActivityFilters([...activityFilters, activityTitle]);
        setTempActivityFilter('');
      }
    }
  };

  const handleRemoveNameFilter = (filterToRemove: string): void => {
    setNameFilters(nameFilters.filter((filter) => filter !== filterToRemove));
  };

  const handleRemoveActivityFilter = (filterToRemove: string): void => {
    setActivityFilters(
      activityFilters.filter((filter) => filter !== filterToRemove)
    );
  };

  const acc = {};
  const userAcc = {};
  const activityAcc = {};
  const usersPerActivityCounter = {};
  const uniqueAnswersWithUser: [SubmissionAnswer, User][] = [];
  const uniqueUsers: User[] = [];
  const activitiesPerUserCounter = {};

  submissions.forEach((submission) => {
    submission.answers.forEach((answer) => {
      const key = `${answer.activity.title.toLowerCase()}|${submission.user.name.toLowerCase()}`;
      if (!acc[key]) {
        acc[key] = true;
        uniqueAnswersWithUser.push([answer, submission.user]);
      }
      const userKey = submission.user.id;
      if (!userAcc[userKey]) {
        userAcc[userKey] = true;
        uniqueUsers.push(submission.user);
      }
      if (!usersPerActivityCounter[answer.activity.id]) {
        usersPerActivityCounter[answer.activity.id] = 0;
      }
      usersPerActivityCounter[answer.activity.id]++;

      if (!activitiesPerUserCounter[userKey]) {
        activitiesPerUserCounter[userKey] = 0;
      }
      activitiesPerUserCounter[userKey]++;
    });
  });

  const filteredUniqueSubmissions = uniqueAnswersWithUser.filter(
    (uniqueAnswer) => {
      const [answer, user] = uniqueAnswer;
      return (
        (nameFilters.length === 0 ||
          nameFilters.some((filter) =>
            user.name.toLowerCase().includes(filter.toLowerCase())
          )) &&
        (activityFilters.length === 0 ||
          activityFilters.some((filter) =>
            answer.activity.title.toLowerCase().includes(filter.toLowerCase())
          ))
      );
    }
  );

  const uniqueUserSet = new Set<User>();
  const uniqueActivityMap = new Map();
  filteredUniqueSubmissions.forEach(([answer, user]) => {
    uniqueUserSet.add(user);
    if (!uniqueActivityMap.has(answer.activity.id)) {
      uniqueActivityMap.set(answer.activity.id, answer.activity);
    }
  });
  const uniqueUserArray = Array.from(uniqueUserSet);
  const uniqueActivityArray = Array.from(uniqueActivityMap.values());

  return (
    <div className="AssembleCrew__background">
      {showCrewModal && (
        <Modal size={ModalSize.TOAST}>
          <p>Crew phone numbers have been copied to your clipboard.</p>
        </Modal>
      )}
      <Modal size={ModalSize.FULLSCREEN}>
        <div className="flex flex-col gap-4 justify-between">
          <div>
            <div className="AssembleCrew__titleContainer">
              <h1 className="AssembleCrew__title text-left">
                Assemble your crew!
              </h1>
            </div>
            <p className="text-left text-sm/4">
              Select the types of adventures that you want to go on and/or the
              people you want to go with. When you're satsified with your crew,
              click 'Assemble Crew' to copy the phone numbers into your
              clipboard.
            </p>
            <div className="w-full mt-4">
              <Field>
                <Label>Adventures</Label>
                <Combobox onChange={handleAddActivityFilter}>
                  <div className="relative w-full">
                    <ComboboxInput
                      className={clsx(
                        'w-full rounded-lg border border-black bg-white py-1.5 pr-8 pl-3 text-sm text-black',
                        'focus:outline-none focus:ring-2 focus:ring-black'
                      )}
                      onChange={(event) =>
                        setActivityComboboxQuery(event.target.value)
                      }
                    />
                    <ComboboxButton className="group absolute inset-y-0 right-0 px-2.5">
                      <ChevronDownIcon className="h-5 w-5 text-black" />
                    </ComboboxButton>
                  </div>
                  <Transition
                    leave="transition ease-in duration-100"
                    leaveFrom="opacity-100"
                    leaveTo="opacity-0"
                    afterLeave={() => setActivityComboboxQuery('')}
                  >
                    <ComboboxOptions
                      anchor="bottom"
                      className="absolute mt-1 max-h-60 w-full overflow-auto rounded-lg border border-black bg-white p-1 shadow-lg AssembleCrew__popup"
                    >
                      {uniqueAnswersWithUser
                        .filter(
                          ([uniqueAnswer]) =>
                            activityFilters.length === 0 ||
                            !activityFilters.some((filter) =>
                              uniqueAnswer.activity.title
                                .toLowerCase()
                                .includes(filter.toLowerCase())
                            )
                        )
                        .filter(
                          ([uniqueAnswer]) =>
                            activityComboboxQuery.length === 0 ||
                            uniqueAnswer.activity.title
                              .toLowerCase()
                              .includes(activityComboboxQuery.toLowerCase())
                        )
                        .map(([uniqueAnswer]) => (
                          <ComboboxOption
                            key={uniqueAnswer.id}
                            value={uniqueAnswer}
                            className="group flex cursor-default items-center gap-2 rounded-lg py-1.5 px-3 select-none hover:bg-gray-100 focus:bg-gray-200"
                          >
                            <CheckIcon className="h-5 w-5 text-black group-focus-within:block hidden" />
                            <div className="text-sm text-black">
                              {uniqueAnswer.activity.title}
                            </div>
                          </ComboboxOption>
                        ))}
                    </ComboboxOptions>
                  </Transition>
                </Combobox>
              </Field>
              <Field>
                <Label>People</Label>
                <Combobox onChange={handleAddNameFilter}>
                  <div className="relative w-full">
                    <ComboboxInput
                      className={clsx(
                        'w-full rounded-lg border border-black bg-white py-1.5 pr-8 pl-3 text-sm text-black',
                        'focus:outline-none focus:ring-2 focus:ring-black'
                      )}
                      onChange={(event) =>
                        setNameComboboxQuery(event.target.value)
                      }
                    />
                    <ComboboxButton className="group absolute inset-y-0 right-0 px-2.5">
                      <ChevronDownIcon className="h-5 w-5 text-black" />
                    </ComboboxButton>
                  </div>
                  <Transition
                    leave="transition ease-in duration-100"
                    leaveFrom="opacity-100"
                    leaveTo="opacity-0"
                    afterLeave={() => setNameComboboxQuery('')}
                  >
                    <ComboboxOptions
                      anchor="bottom"
                      className="absolute mt-1 max-h-60 w-full overflow-auto rounded-lg border border-black bg-white p-1 shadow-lg AssembleCrew__popup"
                    >
                      {uniqueUsers
                        .filter(
                          (uniqueUser) =>
                            nameFilters.length === 0 ||
                            !nameFilters.some((filter) =>
                              uniqueUser.name
                                .toLowerCase()
                                .includes(filter.toLowerCase())
                            )
                        )
                        .filter(
                          (uniqueUser) =>
                            nameComboboxQuery.length === 0 ||
                            uniqueUser.name
                              .toLowerCase()
                              .includes(nameComboboxQuery.toLowerCase())
                        )
                        .map((uniqueUser) => (
                          <ComboboxOption
                            key={uniqueUser.id}
                            value={uniqueUser}
                            className="group flex cursor-default items-center gap-2 rounded-lg py-1.5 px-3 select-none hover:bg-gray-100 focus:bg-gray-200"
                          >
                            <CheckIcon className="h-5 w-5 text-black group-focus-within:block hidden" />
                            <div className="text-sm text-black">
                              {uniqueUser.name}
                            </div>
                          </ComboboxOption>
                        ))}
                    </ComboboxOptions>
                  </Transition>
                </Combobox>
              </Field>
            </div>
            <Field className="mt-4 mb-4 w-full">
              <Label>Filters</Label>
              <div className="flex flex-row gap-1 flex-wrap flex-start w-full mt-2 mb-8">
                {[
                  ...nameFilters.map((filter) => ({
                    label: filter,
                    type: ChipType.PERSON,
                  })),
                  ...activityFilters.map((filter) => ({
                    label: filter,
                    type: ChipType.ACTIVITY,
                  })),
                ].map((filter, index) => (
                  <Chip
                    key={index}
                    label={filter.label}
                    onDelete={() => {
                      handleRemoveNameFilter(filter.label);
                      handleRemoveActivityFilter(filter.label);
                    }}
                    type={filter.type}
                  />
                ))}
              </div>
            </Field>

            <TabGroup className="w-full">
              <TabList className="flex gap-4 w-full">
                <Tab
                  key="people"
                  className="rounded-full py-1 px-3 text-sm/6 font-semibold text-black focus:outline-none data-[selected]:bg-black/10 data-[hover]:bg-black/5 data-[selected]:data-[hover]:bg-black/10 data-[focus]:outline-1 data-[focus]:outline-black"
                >
                  Crew
                </Tab>
                <Tab
                  key="activities"
                  className="rounded-full py-1 px-3 text-sm/6 font-semibold text-black focus:outline-none data-[selected]:bg-black/10 data-[hover]:bg-black/5 data-[selected]:data-[hover]:bg-black/10 data-[focus]:outline-1 data-[focus]:outline-black"
                >
                  Adventures
                </Tab>
              </TabList>
              <TabPanels className="mt-3 w-full">
                <TabPanel key="people" className="rounded-xl bg-black/5 p-3">
                  <ul>
                    {uniqueUserArray.map((user) => (
                      <PersonListItem
                        key={user.id}
                        user={user}
                        onClick={(u) => {
                          if (
                            !notSelectedUsers.some(
                              (selectedUser) => selectedUser.id === u.id
                            )
                          ) {
                            setNotSelectedUsers([...notSelectedUsers, u]);
                          } else {
                            setNotSelectedUsers(
                              notSelectedUsers.filter(
                                (selectedUser) => selectedUser.id !== u.id
                              )
                            );
                          }
                        }}
                        actionArea={() => <input
                          type="checkbox"
                          className="ml-2 h-4 w-4 align-middle"
                          readOnly
                          checked={
                            !notSelectedUsers.some(
                              (selectedUser) => selectedUser.id === user.id
                            )
                          }
                        />}
                      />
                    ))}
                    {uniqueUserArray.length === 0 && (
                      <li className="relative rounded-md p-3 text-sm/6 transition hover:bg-black/5 text-left flex items-center justify-between">
                        No potential crew members
                      </li>
                    )}
                  </ul>
                </TabPanel>
                <TabPanel
                  key="activities"
                  className="rounded-xl bg-black/5 p-3"
                >
                  <ul>
                    {uniqueActivityArray.map((activity) => (
                      <li
                        key={activity.id}
                        className="relative rounded-md p-3 text-sm/6 transition hover:bg-black/5 text-left"
                      >
                        <a
                          href="#"
                          className="font-semibold text-black text-left"
                        >
                          <span className="absolute inset-0" />
                          {activity.title}
                        </a>
                        <ul
                          className="flex gap-2 text-black/50"
                          aria-hidden="true"
                        >
                          {categories ? (
                            <li>
                              {categories &&
                                categories.find(
                                  (category) =>
                                    category.id === activity.categoryId
                                )?.title}
                            </li>
                          ) : null}
                          <li aria-hidden="true">&middot;</li>
                          <li>{usersPerActivityCounter[activity.id]} people</li>
                        </ul>
                      </li>
                    ))}
                    {filteredUniqueSubmissions.length === 0 && (
                      <li className="relative rounded-md p-3 text-sm/6 transition hover:bg-black/5 text-left flex items-center justify-between">
                        No overlapping adventures
                      </li>
                    )}
                  </ul>
                </TabPanel>
              </TabPanels>
            </TabGroup>
          </div>
          <div>
            <Button
              title="Assemble Crew"
              onClick={() => {
                const selectedUsersPhoneNumbers = uniqueUserArray
                  .filter(
                    (user) =>
                      !notSelectedUsers.some(
                        (selectedUser) => selectedUser.id === user.id
                      )
                  )
                  .map((user) => user.phoneNumber);
                navigator.clipboard.writeText(
                  selectedUsersPhoneNumbers.join(', ')
                );
                setShowCrewModal(true);
                setTimeout(() => {
                  setShowCrewModal(false);
                }, 3000);
              }}
            />
          </div>
        </div>
      </Modal>
    </div>
  );
};
