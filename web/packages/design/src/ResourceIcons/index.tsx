import React, { ComponentProps } from 'react';
import { useTheme } from 'styled-components';
import { Image } from 'design';

import application from './assets/application.svg';
import server from './assets/server.svg';
import windowsDark from './assets/windows-dark.svg';
import windowsLight from './assets/windows-light.svg';

/** Uses given icon for all themes. */
const forAllThemes = icon => ({ dark: icon, light: icon });

/** A name->theme->spec mapping of resource icons. */
const iconSpecs = {
  Application: forAllThemes(application),
  Server: forAllThemes(server),
  Windows: { dark: windowsDark, light: windowsLight },
};

export type ResourceIconName = keyof typeof iconSpecs;

type ImageProps = ComponentProps<typeof Image>;
interface ThemedIconProps extends ImageProps {
  name: ResourceIconName;
}

/** Displays an icon of a given name for current theme. */
const ThemedIcon = ({ name, ...props }: ThemedIconProps) => {
  const theme = useTheme();
  const icon = iconSpecs[name][theme.name];
  return <Image src={icon} {...props} />;
};

type IconsMap = {
  [name in ResourceIconName]: React.FC<ImageProps>;
};

/**
 * Maps icon name to a component that displays a theme-aware resource icon.  The
 * icon component exposes props of the underlying `Image`.
 */
export const resourceIcons = Object.fromEntries(
  Object.entries(iconSpecs).map(([name, icon]) => [
    name,

    ({ ...props }: ImageProps) => <ThemedIcon name={name} {...props} />,
  ])
  // Ugly type cast required because TS can't guarantee precise types for
  // Object.entries() and Object.fromEntries().
) as IconsMap;
