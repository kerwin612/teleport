import React from 'react';
import { Meta, StoryObj } from '@storybook/react';
import { ResourceCard as ResourceCard } from './ResourceCard';
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
        {[...apps, ...databases].map(res => (
          <ResourceCard resource={res} />
        ))}
      </>
    );
  },
};
