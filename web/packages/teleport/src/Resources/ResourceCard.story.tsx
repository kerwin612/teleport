import React from 'react';
import { Meta, StoryObj } from '@storybook/react';
import { ResourceCard as ResourceCard } from './ResourceCard';
import { Node } from 'teleport/services/nodes/types';
import { Database } from 'teleport/services/databases/types';
import { App } from 'teleport/services/apps';
import makeApp from 'teleport/services/apps/makeApps';
import { apps } from 'teleport/Apps/fixtures';
import { databases } from 'teleport/Databases/fixtures';

const meta: Meta<typeof ResourceCard> = {
  component: ResourceCard,
  title: 'Teleport/Resources/ResourceCard',
};

export default meta;
type Story = StoryObj<typeof ResourceCard>;

export const Cards: Story = {
  render() {
    return (
      <>
        {[...apps, ...databases].map(app => (
          <ResourceCard resource={app} />
        ))}
      </>
    );
  },
};

// export const DatabaseCard: Story = {
//   args: {
//     resource: {
//       name: 'database-1',
//     } as Database,
//   },
// };
