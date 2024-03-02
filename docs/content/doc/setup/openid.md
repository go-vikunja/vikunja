# OpenID

Vikunja allows for authentication with an oauth provider via the OpenID standard.

To learn more about how to configure this, [check out the examples]({{< ref "openid-examples.md">}})

{{< table_of_contents >}}

## Automatically assign users to teams

Vikunja is capable of automatically adding users to a team based on a group defined in the oidc provider.
If configured, Vikunja will sync teams, automatically create new ones and make sure the members are part of the configured teams.
Teams which exist only because they were created from oidc attributes are not editable in Vikunja.

To distinguish between teams created in Vikunja and teams generated automatically via oidc, generated teams have an `oidcID` assigned internally.

You need to make sure the OpenID provider offers a `vikunja_groups` key through your custom scope. This is the key, which is looked up by Vikunja to start the procedure.

Additionally, make sure to deliver an `oidcID` and a `name` attribute within the `vikunja_groups`. You can see how to set this up, if you continue reading.

### Setup in Authentik

To configure automatic team management through Authentik, we assume you have already [set up Authentik]({{< ref "openid-examples.md">}}#authentik) as an oidc provider for authentication with Vikunja.

To use Authentik's group assignment feature, follow these steps:

1. Edit [your config]({{< ref "config.md">}}) to include the following scopes: `openid profile email vikunja_scope`
2. Open `<your authentik url>/if/admin/#/core/property-mappings`
3. Create a new property mapping called `vikunja_scope` as scope mapping. There is a field `expression` to enter python expressions that will be delivered with the oidc token.
4. Write a small script like the following to add group information to `vikunja_scope`:

```python
groupsDict = {"vikunja_groups": []}
for group in request.user.ak_groups.all():
  groupsDict["vikunja_groups"].append({"name": group.name, "oidcID": group.num_pk})
return groupsDict
```

output example:

```
{
    "vikunja_groups": [
        {
            "name": "team 1",
            "oidcID": 33349
        },
        {
            "name": "team 2",
            "oidcID": 35933
        }
    ]
}
```

5. In Authentik's menu on the left, go to Applications > Providers > Select the Vikunja provider. Then click on "Edit", on the bottom open "Advanced protocol settings", select the newly created property mapping under "Scopes". Save the provider.

Now when you log into Vikunja via Authentik it will show you a list of scopes you are claiming.
You should see the description you entered on the oidc provider's admin area.

Proceed to vikunja and open the teams page in the sidebar menu.
You should see "(sso: *your_oidcID*)" written next to each team you were assigned through oidc.

## Setup in Keycloak

The kind people from the Darmstadt Makerspace have written [a guide on how to create a mapper for Vikunja here](https://github.com/makerspace-darmstadt/keycloak-vikunja-mapper).

## Use cases

All examples assume one team called "Team 1" in your provider.

* *Token delivers team.name +team.oidcID and Vikunja team does not exist:* \
New team will be created called "Team 1" with attribute oidcID: "33929"

2. *In Vikunja Team with name "team 1" already exists in vikunja, but has no oidcID set:* \
new team will be created called "team 1" with attribute oidcID: "33929"

3. *In Vikunja Team with name "team 1" already exists in vikunja, but has different oidcID set:* \
new team will be created called "team 1" with attribute oidcID: "33929"

4. *In Vikunja Team with oidcID "33929" already exists in vikunja, but has different name than "team1":* \
new team will be created called "team 1" with attribute oidcID: "33929"

5. *Scope vikunja_scope is not set:* \
nothing happens

6. *oidcID is not set:* \
You'll get error.
Custom Scope malformed
"The custom scope set by the OIDC provider is malformed. Please make sure the openid provider sets the data correctly for your scope. Check especially to have set an oidcID."

7. *In Vikunja I am in "team 3" with oidcID "", but the token does not deliver any data for "team 3":* \
You will stay in team 3 since it was not set by the oidc provider

8. *In Vikunja I am in "team 3" with oidcID "12345", but the token does not deliver any data for "team 3"*:\
You will be signed out of all teams, which have an oidcID set and are not contained in the token.
Especially if you've been the last team member, the team will be deleted.
