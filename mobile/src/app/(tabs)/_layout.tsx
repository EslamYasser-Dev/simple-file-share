import { Tabs } from 'expo-router';
import { Text } from 'react-native';
import { colors } from '../_layout';

function TabIcon({ symbol, focused }: { symbol: string; focused: boolean }) {
  return (
    <Text style={{ fontSize: 18, color: focused ? colors.accent : colors.muted }}>{symbol}</Text>
  );
}

export default function TabsLayout() {
  return (
    <Tabs
      screenOptions={{
        headerStyle: { backgroundColor: colors.background },
        headerTintColor: colors.text,
        tabBarStyle: { backgroundColor: colors.card, borderTopColor: colors.border },
        tabBarActiveTintColor: colors.accent,
        tabBarInactiveTintColor: colors.muted,
      }}
    >
      <Tabs.Screen
        name="index"
        options={{
          title: 'Files',
          tabBarIcon: ({ focused }) => <TabIcon symbol="📁" focused={focused} />,
        }}
      />
      <Tabs.Screen
        name="shares"
        options={{
          title: 'Shares',
          tabBarIcon: ({ focused }) => <TabIcon symbol="🔗" focused={focused} />,
        }}
      />
      <Tabs.Screen
        name="account"
        options={{
          title: 'Account',
          tabBarIcon: ({ focused }) => <TabIcon symbol="👤" focused={focused} />,
        }}
      />
    </Tabs>
  );
}
