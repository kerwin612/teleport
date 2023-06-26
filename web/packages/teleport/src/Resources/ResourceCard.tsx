import React from 'react';
import styled from 'styled-components';

import { Flex } from 'design';

import { AgentKind } from 'teleport/services/agents';

type Props = {
  resource: AgentKind;
};
export const ResourceCard = ({ resource }: Props) => {
  return (
    <CardContainer key={resource.name}>
      <div>{resource.name}</div>
    </CardContainer>
  );
};

export const CardContainer = styled(Flex)`
  border-top: 2px solid ${props => props.theme.colors.spotBackground[0]};
  padding: ${props => props.theme.space[3]}px;

  @media (min-width: ${props => props.theme.breakpoints.tablet}px) {
    border: ${props => props.theme.borders[2]}
      ${props => props.theme.colors.spotBackground[0]};
    border-radius: ${props => props.theme.radii[3]}px;
  }
`;
