import React from 'react';
import styled from 'styled-components';

import { Button, Flex, Pill, Text } from 'design';

import { AgentKind } from 'teleport/services/agents';
import { CheckboxInput } from 'design/Checkbox';
import { resourceIcons } from 'design/ResourceIcons';
import * as Icons from 'design/Icon';
import { Database } from 'teleport/services/databases';

const ResourceDetails = styled(Flex)`
  flex-grow: 1;
`;

const ResourceName = styled.div`
  flex-grow: 1;
`;

type Props = {
  resource: AgentKind;
};
export const ResourceCard = ({ resource }: Props) => {
  const ResIcon = agentIcon(resource);
  return (
    <CardContainer p={3} alignItems="start">
      <CheckboxInput type="checkbox"></CheckboxInput>
      <ResIcon width="60px" height="60px" />
      <ResourceDetails flexDirection="column">
        <Flex flexDirection="row">
          <ResourceName>{agentName(resource)}</ResourceName>
          <Button>Connect</Button>
        </Flex>
        <Flex flexDirection="row">
          <Icons.Database />
          <div>{agentType(resource)}</div>
          <div>{resource.description}</div>
        </Flex>
        <div>
          {resource.labels.map(({ name, value }) => (
            <Pill label={`${name}: ${value}`} />
          ))}
        </div>
      </ResourceDetails>
    </CardContainer>
  );
};

function agentType(agent: AgentKind) {
  return 'type' in agent ? agent.type : '';
}

function agentName(agent: AgentKind) {
  return agent.kind === 'node' ? agent.hostname : agent.name;
}

function agentIcon(agent: AgentKind) {
  switch (agent.kind) {
    case 'app':
      return resourceIcons.Application;
    case 'db':
      return resourceIcons.Server;
    default:
      return resourceIcons.Server;
  }
}

function databaseIcon(db: Database) {}

export const CardContainer = styled(Flex)`
  border-top: 2px solid ${props => props.theme.colors.spotBackground[0]};

  @media (min-width: ${props => props.theme.breakpoints.tablet}px) {
    border: ${props => props.theme.borders[2]}
      ${props => props.theme.colors.spotBackground[0]};
    border-radius: ${props => props.theme.radii[3]}px;
  }
`;
