import * as Headless from '@headlessui/react'
import React, { forwardRef } from 'react'
import { Link as TLink} from '@tanstack/react-router'

export const Link = forwardRef(function Link(
  props: { href: string } & React.ComponentPropsWithoutRef<'a'>,
  ref: React.ForwardedRef<HTMLAnchorElement>
) {
  return (
    <Headless.DataInteractive>
      <TLink {...props} to={props.href} ref={ref} />
    </Headless.DataInteractive>
  )
})
