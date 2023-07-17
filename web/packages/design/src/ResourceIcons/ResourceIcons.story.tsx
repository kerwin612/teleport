import React from 'react';
import { resourceIcons } from 'design/ResourceIcons';
import { Flex } from '..';

export default {
  title: 'Design/ResourceIcons',
};

export const Icons = () => {
  return (
    <>
      {Object.entries(resourceIcons).map(([name, Icon]) => (
        <Flex gap={3} alignItems="center">
          <Icon width="100px" /> <Icon width="50px" />
          {name}
        </Flex>
      ))}
    </>
  );
};
