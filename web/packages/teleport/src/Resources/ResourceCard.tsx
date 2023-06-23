import React from 'react';
import styled from 'styled-components';

import { Flex } from 'design';

import { UnifiedResource } from 'teleport/services/resources';

type Props = {
  resource: UnifiedResource;
};
export const ResourceCard = ({ resource }: Props) => {
  return (
    <CardContainer key={resource.name}>
      <div>hello</div>
    </CardContainer>
  );
};

export const CardContainer = styled(Flex)`
  border: 2px solid grey;
  padding: 16px;
`;
